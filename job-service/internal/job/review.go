package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
	"job-service/internal/conf"
)

// 消息结构体 (canal 格式)
type Msg struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	Table    string `json:"table"`
	IsDdl    bool   `json:"isDdl"`
	Data     []map[string]interface{}
}

// ES 客户端封装
type EsClient struct {
	*elasticsearch.TypedClient
}

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// 作业 Worker,用于消费 Kafka 消息并同步至 Elasticsearch
type JobWork struct {
	kafkaReader   *kafka.Reader
	esClient      *EsClient
	topicIndexMap map[string]string
	log           *log.Helper
}

func NewJobWrok(kafkaReader *kafka.Reader, esClient *EsClient, conf *conf.Elasticsearch, logger log.Logger) *JobWork {
	topicIndexMap := make(map[string]string)
	for _, idx := range conf.Indices {
		topicIndexMap[idx.Topic] = idx.Index
	}
	return &JobWork{
		kafkaReader:   kafkaReader,
		esClient:      esClient,
		topicIndexMap: topicIndexMap,
		log:           log.NewHelper(logger),
	}
}

// kafka
// Kafka Reader
func NewKafkaReader(c *conf.Kafka) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.Brokers,
		GroupTopics: c.Topics,
		GroupID:     c.GroupId,
	})
}

// elsaticsearch
func NewESClient(conf *conf.Elasticsearch) (*EsClient, error) {
	// ES 配置
	c := elasticsearch.Config{Addresses: conf.Addresses}

	// 创建客户端连接
	client, err := elasticsearch.NewTypedClient(c)
	if err != nil {
		return nil, err
	}

	return &EsClient{
		TypedClient: client,
	}, nil
}

// 启动消费循环，将 canal->kafka 的变更消息同步到 Elasticsearch
func (jw JobWork) Start(ctx context.Context) error {
	jw.log.WithContext(ctx).Info("job work start")

	// 1. 从kafka中获取MySQL中的数据变更消息
	// 接收消息
	for {
		// 读取 Kafka 消息
		m, err := jw.kafkaReader.ReadMessage(ctx)
		// 如果上层 ctx 被取消，优雅退出
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if err != nil {
			jw.log.Errorf("read message failed:%v\n", err)
			break
		}
		jw.log.WithContext(ctx).Infof("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))

		// 根据当前 topic 查找对应索引
		index, ok := jw.topicIndexMap[m.Topic]
		if !ok {
			jw.log.WithContext(ctx).Errorf("no index mapping for topic: %s", m.Topic)
			continue
		}

		// 反序列化 canal 消息
		msg := new(Msg)
		if err := json.Unmarshal(m.Value, msg); err != nil {
			jw.log.WithContext(ctx).Errorf("unmarshal message failed:%v\n", err)
			continue
		}

		// 遍历变更的行数据
		for _, data := range msg.Data {
			// 提取唯一 ID（用于 ES 的文档 _id）
			docID := jw.extractID(data)
			if docID == "" {
				jw.log.WithContext(ctx).Error("missing id in message, skipping")
				continue
			}

			// 根据 canal 类型选择插入或更新
			switch msg.Type {
			case "INSERT":
				jw.indexDocument(ctx, index, docID, data)
			case "UPDATE":
				jw.updateDocument(ctx, index, docID, data)
			default:
				jw.log.WithContext(ctx).Infof("unsupported message type: %s, skipping", msg.Type)
			}
		}
	}
	return nil
}

// // 提取唯一 id，用于 ES 写入/更新时做文档 _id，便于幂等写入
func (jw *JobWork) extractID(data map[string]interface{}) string {
	if id, ok := data["id"]; ok {
		switch v := id.(type) {
		case string:
			return v
		case float64:
			// Canal 转换时可能是 float64，转回字符串
			return fmt.Sprintf("%.0f", v)
		}
	}
	return ""
}

// 在 Elasticsearch 中插入文档（幂等写入）
func (jw *JobWork) indexDocument(ctx context.Context, index, id string, data map[string]interface{}) {
	_, err := jw.esClient.Index(index).Id(id).Document(data).Do(ctx)
	if err != nil {
		jw.log.WithContext(ctx).Errorf("index document failed: %v", err)
	} else {
		jw.log.WithContext(ctx).Infof("indexed document id=%s into index=%s", id, index)
	}
}

// 在 Elasticsearch 中更新文档（幂等更新）
func (jw *JobWork) updateDocument(ctx context.Context, index, id string, data map[string]interface{}) {
	_, err := jw.esClient.Update(index, id).Doc(data).DocAsUpsert(true).Do(ctx)
	if err != nil {
		jw.log.WithContext(ctx).Errorf("update document failed: %v", err)
	} else {
		jw.log.WithContext(ctx).Infof("updated document id=%s in index=%s", id, index)
	}
}

func (jw JobWork) Stop(ctx context.Context) error {
	jw.log.WithContext(ctx).Info("job work stop")
	return jw.kafkaReader.Close()
}
