# 短视频项目

## 一、概述

1.1 项目简介

使用kratos框架构建短视频后端平台，实现用户注册、登录，视频上传，获取视频流，视频搜索，点赞，评论，关注等功能。

部署到k8s集群，基于peometheus+grafnan监控

<img width="2214" height="1132" alt="image" src="https://github.com/user-attachments/assets/3b887f7d-aeef-434e-87e9-37edbe39d0ac" />


1.2 项目运行

```go
// 下载项目
git clone git@github.com:yong-king/tiktok.git
// 安装依赖
go mod tidy
// 运行
cd /bin
xxxx-service.exe

// 编译
make build
```



## 二、项目结构

|------------------   comment-service （评论服务）

|------------------  favorite-service (点赞服务)

|------------------  feed-service （视频流服务）

|------------------  job-service （消息同步服务）

|------------------  relation-service （用户粉丝关系服务）

|------------------  user-service （用户服务）

|------------------  video-service （视频服务）

|------------------ docker (docker 配置文件)

|------------------ k8s （k8s配置文件）

|------------------ docker-compose

## 三、基础接口列表（xxx-service/api/xxx/v1/xxx.proto）

1. user-service(http: 8081, grpc: 9081)

   ```go
   // 用户注册
   POST 1270.0.0.1:8081/api/user/register
   // 用户登录
   POST 1270.0.0.1:8081/api/user/login
   ```

2. video-service(8080，8082, 9082)

   ```go
   // 视频上传
   POST 1270.0.0.1:8080/api/video/upload
   
   // 获取视频列表
   GET 127.0.0.1:8082/api/video
   
   // 视频信息
   POST 127.0.0.1:8082/api/video/creat
   ```

3. feed-service(8085, 9085)

   ```go
   GET 127.0.0.1:8085/api/feed
   ```

4. comment-service(8084, 9084)

   ```go
   // 发表评论
   POST 127.0.0.1:8084/api/comment/creat
   
   // 获取评论
   GET 127.0.0.1:8084/api/comment/get
   ```

5. favorite-service(8083, 9083)

   ```go
   // 点赞
   POST 127.0.0.1:8083/api/favorite/action
   
   // 获取点赞视频列表
   GET 127.0.0.1:8083/api/favorite/videos
   ```

6. relation-service (8086, 9086)

   ```go
   // 关注
   POST 127.0.0.1:8086/api/relation/control
   
   // 获取关注列表
   GET 127.0.0.1:8086/api/relation/list
   ```

## 四、组件

1. 数据库MySQL（3306），redis（6379）
2. 文件服务器minio（9001）
3. 注册中心consul
4. 数据读写分离
   1. canal
   2. kafka
   3. elastic search
      1. kibana（可视化）

5. 链路追踪
   1. openTelemetry
   2. jaeger（可视化）

6. 监控
   1. Prometheus
   2. grafana

7. zap日志库，压测wrk，jwt，docker，k8s，令牌桶限流，熔断

   

## 五、框架

<img width="1862" height="1115" alt="image" src="https://github.com/user-attachments/assets/5a89d753-1b8d-4e9e-b478-84df60ab3fb7" />


1. user微服务

   | 注册                                                         | 登录                                                         | 用户信息获取                                                 |
   | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
   | ![注册接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E6%B3%A8%E5%86%8C%E6%8E%A5%E5%8F%A3.png) | ![登录接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E7%99%BB%E5%BD%95%E6%8E%A5%E5%8F%A3.png) | ![获取用户信息接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E4%BF%A1%E6%81%AF%E6%8E%A5%E5%8F%A3.png) |

   

2. feed

   | 获取视频流                                                   |
   | ------------------------------------------------------------ |
   | ![获取视频流接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E8%A7%86%E9%A2%91%E6%B5%81%E6%8E%A5%E5%8F%A3.png) |

   

3. video

   | 上传视频                                                     | 获取视频                                                     |
   | ------------------------------------------------------------ | ------------------------------------------------------------ |
   | ![投稿视频接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E6%8A%95%E7%A8%BF%E8%A7%86%E9%A2%91%E6%8E%A5%E5%8F%A3.png) | ![获取用户发布视频列表接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E5%8F%91%E5%B8%83%E8%A7%86%E9%A2%91%E5%88%97%E8%A1%A8%E6%8E%A5%E5%8F%A3.png) |

   

4. comment

   | 评论操作                                                     | 获取评论                                                     |
   | ------------------------------------------------------------ | ------------------------------------------------------------ |
   | ![评论操作接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%AF%84%E8%AE%BA%E6%93%8D%E4%BD%9C%E6%8E%A5%E5%8F%A3.png) | ![获取视频评论列表接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E8%A7%86%E9%A2%91%E8%AF%84%E8%AE%BA%E5%88%97%E8%A1%A8%E6%8E%A5%E5%8F%A3.png) |

   

5. favorite

   | 点赞                                                         | 获取点赞                                                     |
   | ------------------------------------------------------------ | ------------------------------------------------------------ |
   | ![点赞操作接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E7%82%B9%E8%B5%9E%E6%93%8D%E4%BD%9C%E6%8E%A5%E5%8F%A3.png) | ![获取用户点赞列表接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E7%94%A8%E6%88%B7%E7%82%B9%E8%B5%9E%E5%88%97%E8%A1%A8%E6%8E%A5%E5%8F%A3.png) |

   

6. relation

   | 关注                                                         | 获取关注列表                                                 |
   | ------------------------------------------------------------ | ------------------------------------------------------------ |
   | ![关注操作接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E5%85%B3%E6%B3%A8%E6%93%8D%E4%BD%9C%E6%8E%A5%E5%8F%A3.png) | ![获取关注列表接口.png](https://github.com/a76yyyy/tiktok/raw/main/pic/%E8%8E%B7%E5%8F%96%E5%85%B3%E6%B3%A8%E5%88%97%E8%A1%A8%E6%8E%A5%E5%8F%A3.png) |

   ## 

## Note

项目架构来自于第三届字节青训营的内容](https://github.com/a76yyyy/tiktok)	

开发具体流程件Readme文件夹
