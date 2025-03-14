# 考前安全密码分发系统流程

## 1. 系统启动

- 启动服务后，访问 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) 可以查看所有 API 接口文档。
- 系统使用 SQLite 数据库，会在项目根目录生成 `exam_passwords.db` 文件。
- 系统首次启动时会自动检测是否存在顶级管理员；如果没有，则采用默认配置自动创建顶级管理员。

## 2. 密码生成与分发

- 管理员通过 **POST /generatePasswords** 接口批量生成密码，可通过请求中指定 `role`（默认值为 "student"；生成管理员密码时需指定为 "admin"）。

```json
请求示例（生成考生密码）：
{
    "count": 10,
    "examid": "EXAM001",
    "sttime": "2024-02-17T15:30:45Z",
    "endtime": "2024-02-17T16:30:45Z"
}

请求示例（生成管理员密码）：
{
  "count": 1,
  "role": "admin",
  "examid": "EXAM001",
  "sttime": "2024-02-17T15:30:45Z",
  "endtime": "2024-02-17T16:30:45Z"
}

返回示例：
{
    "message": "生成密码数量：10"
}
```

## 3. 密码获取

### 3.1 考生获取密码

- 考生通过 **GET /password?stuid=学号** 接口领取未分发的考试密码。

```json
请求示例：
GET /password?stuid=2021001

返回示例：
{
    "password": "a1b2c3d4"
}
```

### 3.2 管理员获取密码

- 管理员通过 **GET /admin/password?examid=考试 ID** 接口领取管理员密码。

```json
请求示例：
GET /admin/password?examid=EXAM001

返回示例：
{
    "examid": "EXAM001",
    "adminid": "admin001",
    "password": "a1b2c3d4"
}

错误示例：
{
    "error": "无可用的管理员密码"
}
```

## 4. 考试入口验证

- 客户端统一使用 **POST /validateExamPassword** 接口验证密码。请求中字段 `stuid` 同时代表学生 ID 或管理员 ID，服务端依角色进行验证。
- 返回数据中包括：
  - `role`：用户角色（"admin" 或 "student"）
  - `is_default`：对于管理员，如果返回为 `true` 表示该管理员还在使用默认配置，需要及时修改
  - 统一返回 `examid` 和 `userid` 字段
  - 返回中会含有 `token` 字段，示例中未标明

```json
请求示例：
{
    "examid": "EXAM001",
    "stuid": "2021001",
    "password": "a1b2c3d4"
}

返回示例（学生）：
{
    "message": "验证成功",
    "role": "student",
    "is_default": false,
    "examid": "EXAM001",
    "userid": "2021001"
}

返回示例（管理员使用默认信息）：
{
    "message": "验证成功",
    "role": "admin",
    "is_default": true,
    "examid": "EXAM001",
    "userid": "admin"
}
```

## 5. 题目管理

系统支持两种题目类型：

1. **单选题（single_choice）**：从四个选项中选择一个正确答案
2. **判断题（judgment）**：判断题目的正误，答案为 T（正确）或 F（错误）

### 5.1 创建题目

- 管理员通过 **POST /question/create** 接口创建新题目
- 需要在请求头中携带有效的管理员 token

```json
请求示例（单选题）：
POST /question/create
Authorization: <token>

{
    "examid": "EXAM001",
    "type": "single_choice",
    "content": "以下哪个是正确的？",
    "optionA": "选项A",
    "optionB": "选项B",
    "optionC": "选项C",
    "optionD": "选项D",
    "answer": "A",
    "score": 5
}

请求示例（判断题）：
POST /question/create
Authorization: <token>

{
    "examid": "EXAM001",
    "type": "judgment",
    "content": "这是一个判断题",
    "answer": "T",
    "score": 3
}

返回示例：
{
    "message": "题目创建成功",
    "id": 1,
    "type": "single_choice",
    "examid": "EXAM001"
}
```

### 5.2 获取考试题目列表

- 管理员通过 **GET /question/list** 接口获取指定考试的所有题目
- 需要在请求头中携带有效的管理员 token

```json
请求示例：
GET /question/list?examid=EXAM001
Authorization: <token>

返回示例：
[
    {
        "id": 1,
        "examid": "EXAM001",
        "type": "single_choice",
        "content": "以下哪个是正确的？",
        "optionA": "选项A",
        "optionB": "选项B",
        "optionC": "选项C",
        "optionD": "选项D",
        "answer": "A",
        "score": 5,
        "createAt": "2023-01-01T12:00:00Z"
    },
    {
        "id": 2,
        "examid": "EXAM001",
        "type": "judgment",
        "content": "这是一个判断题",
        "optionA": "",
        "optionB": "",
        "optionC": "",
        "optionD": "",
        "answer": "T",
        "score": 3,
        "createAt": "2023-01-01T12:05:00Z"
    }
]
```

### 5.3 获取题目详情

- 管理员通过 **GET /question/{id}** 接口获取指定题目的详细信息
- 需要在请求头中携带有效的管理员 token

```json
请求示例：
GET /question/1
Authorization: <token>

返回示例：
{
    "id": 1,
    "examid": "EXAM001",
    "type": "single_choice",
    "content": "以下哪个是正确的？",
    "optionA": "选项A",
    "optionB": "选项B",
    "optionC": "选项C",
    "optionD": "选项D",
    "answer": "A",
    "score": 5,
    "createAt": "2023-01-01T12:00:00Z"
}
```

### 5.4 更新题目

- 管理员通过 **PUT /question/{id}** 接口更新指定题目的信息
- 需要在请求头中携带有效的管理员 token

```json
请求示例：
PUT /question/1
Authorization: <token>

{
    "content": "更新后的题干",
    "optionA": "新选项A",
    "optionB": "新选项B",
    "optionC": "新选项C",
    "optionD": "新选项D",
    "answer": "B",
    "score": 10
}

返回示例：
{
    "message": "题目更新成功"
}
```

### 5.5 删除题目

- 管理员通过 **DELETE /question/{id}** 接口删除指定的题目
- 需要在请求头中携带有效的管理员 token

```json
请求示例：
DELETE /question/1
Authorization: <token>

返回示例：
{
    "message": "题目删除成功"
}
```

### 5.6 获取支持的题目类型

- 通过 **GET /question/types** 接口获取系统当前支持的题目类型

```json
返回示例：
{
    "questiontypes": [
        {
            "type": "single_choice",
            "name": "单选题",
            "description": "从四个选项中选择一个正确答案"
        },
        {
            "type": "judgment",
            "name": "判断题",
            "description": "判断题目的正误，答案为T（正确）或F（错误）"
        }
    ]
}
```

## 6. 屏幕获取与直播功能

系统提供两种屏幕直播策略：

1. **定时截屏（interval）**：每隔几秒截取一次屏幕，适合低带宽环境
2. **实时直播（realtime）**：高频率截屏（约 10 帧/秒），提供接近实时的体验

### 6.1 初始化屏幕流会话

- 学生端通过 **POST /screen/init** 接口初始化屏幕流会话
- 需要在请求头中携带有效的 token

```json
请求示例：
POST /screen/init
Authorization: <token>

{
    "streamtype": "interval",  // 可选值: interval, realtime
    "interval": 5              // 截屏间隔秒数（仅用于interval类型）
}

返回示例：
{
    "message": "屏幕流会话初始化成功",
    "sessionid": "abcdef1234567890",
    "userid": "2021001",
    "examid": "EXAM001",
    "streamtype": "interval",
    "interval": 5
}
```

### 6.2 建立 WebSocket 连接

#### 6.2.1 学生端 WebSocket 连接

- 学生端通过 **GET /ws/student** 建立 WebSocket 连接
- 需要提供 userid、examid 和 token 参数

```
WebSocket连接URL示例：
ws://localhost:8080/ws/student?userid=2021001&examid=EXAM001&token=<token>
```

#### 6.2.2 管理员端 WebSocket 连接

- 管理员端通过 **GET /ws/admin** 建立 WebSocket 连接
- 需要提供 userid、examid 和 token 参数

```
WebSocket连接URL示例：
ws://localhost:8080/ws/admin?userid=admin&examid=EXAM001&token=<token>
```

### 6.3 WebSocket 消息格式

#### 6.3.1 开始屏幕共享命令

- 学生端发送开始命令：

```json
{
  "type": "command",
  "command": "start",
  "streamtype": "interval", // 可选值: interval, realtime
  "interval": 5 // 截屏间隔秒数（仅用于interval类型）
}
```

#### 6.3.2 停止屏幕共享命令

- 学生端发送停止命令：

```json
{
  "type": "command",
  "command": "stop"
}
```

#### 6.3.3 屏幕图像消息

- 管理员端接收屏幕图像：

```json
{
  "type": "image",
  "userid": "2021001",
  "examid": "EXAM001",
  "timestamp": 1613558445,
  "data": "base64编码的图像数据",
  "sessionid": "abcdef1234567890"
}
```

#### 6.3.4 状态消息

- 状态确认消息：

```json
{
  "type": "status",
  "userid": "2021001",
  "examid": "EXAM001",
  "timestamp": 1613558445,
  "data": "started", // 或 "stopped"
  "sessionid": "abcdef1234567890"
}
```

### 6.4 获取活跃的屏幕流

- 管理员通过 **GET /screen/active** 接口获取指定考试的所有活跃屏幕流
- 需要在请求头中携带有效的管理员 token

```json
请求示例：
GET /screen/active?examid=EXAM001
Authorization: <token>

返回示例：
[
    {
        "sessionid": "abcdef1234567890",
        "userid": "2021001",
        "username": "张三",
        "examid": "EXAM001",
        "starttime": "2023-01-01T12:00:00Z",
        "streamtype": "interval",
        "interval": 5
    },
    {
        "sessionid": "fedcba0987654321",
        "userid": "2021002",
        "username": "李四",
        "examid": "EXAM001",
        "starttime": "2023-01-01T12:05:00Z",
        "streamtype": "realtime",
        "interval": 0
    }
]
```

### 6.5 停止屏幕流

- 通过 **POST /screen/stop** 接口停止指定的屏幕流会话
- 需要在请求头中携带有效的 token
- 只有会话所有者或管理员可以停止屏幕流

```json
请求示例：
POST /screen/stop
Authorization: <token>

{
    "sessionid": "abcdef1234567890"
}

返回示例：
{
    "message": "屏幕流已停止"
}
```

## 7. 异常情况处理

- 考生可通过 **POST /raiseHand** 接口提交举手请求，反馈领取或验证过程中遇到的问题。

```json
请求示例：
{
    "examid": "EXAM001",
    "stuid": "2021001",
    "reason": "无法获取密码"
}

返回示例：
{
    "message": "举手请求已提交"
}
```

## 8. 管理员相关接口

### 8.1 顶级管理员信息查看

- **GET /admin/backup**：可查看系统初始化时自动创建的顶级管理员备份信息。

```json
请求示例：
GET /admin/backup

返回示例：
{
    "adminid": "admin",
    "password": "admin123",
    "examid": "EXAM001",
    "created_at": "2024-02-17T15:30:45.123Z"
}
```

### 8.2 创建新考试

- **POST /exam/create**：创建新考试，可指定考试 ID 和流类型，如不指定考试 ID 则自动生成。

```json
请求示例：
{
    "examid": "EXAM002",  // 可选，不提供则自动生成
    "streamtype": "interval"  // 可选值: interval, realtime，默认为interval
}

返回示例：
{
    "message": "考试创建成功",
    "examid": "EXAM002",
    "streamtype": "interval"
}
```

### 8.3 更新考试流类型

- **PUT /exam/streamtype**：更新指定考试的屏幕流类型。

```json
请求示例：
{
    "examid": "EXAM001",
    "streamtype": "realtime"  // 可选值: interval, realtime
}

返回示例：
{
    "message": "考试流类型更新成功",
    "examid": "EXAM001",
    "streamtype": "realtime"
}
```

### 8.4 获取支持的流类型

- **GET /exam/streamtypes**：获取系统当前支持的屏幕流类型。

```json
返回示例：
{
    "streamtypes": [
        {
            "type": "interval",
            "name": "定时截屏",
            "description": "每隔几秒截取一次屏幕，适合低带宽环境"
        },
        {
            "type": "realtime",
            "name": "实时直播",
            "description": "高频率截屏（约10帧/秒），提供接近实时的体验"
        }
    ]
}
```

### 8.5 学生绑定考试

- **POST /exam/bind**

```json
请求示例：
{
    "examid": "EXAM001",
    "stuid": "2021001"
}

返回示例：
{
    "message": "绑定成功",
    "examid": "EXAM001",
    "stuid": "2021001",
    "password": "a1b2c3d4"
}
```

### 8.6 更新管理员信息

- 若管理员登录后检测到 `is_default` 为 `true`，说明仍处于默认状态，此时客户端可调用 **POST /admin/update** 接口更新管理员的默认信息（包括管理员 ID、密码和考试 ID）。

```json
请求示例：
{
    "old_adminid": "admin",
    "old_password": "admin123",
    "old_examid": "EXAM001",
    "new_adminid": "customAdmin",
    "new_password": "customPassword",
    "new_examid": "CUSTOM001"
}

返回示例：
{
    "message": "管理员信息更新成功"
}
```

### 8.7 删除管理员数据（测试接口）

- 用于开发测试阶段，**DELETE /admin/deleteAll** 可清除所有管理员数据。

```json
返回示例：
{
    "message": "所有管理员数据已删除"
}
```

## 9. Token 认证机制

### 9.1 Token 的获取与使用

- 所有用户（考生/管理员）通过 **POST /validateExamPassword** 登录接口获取 token
- 获取到的 token 需要在后续请求中通过 `Authorization` 请求头传递
- token 有效期为 24 小时
- token 中包含了用户 ID、考试 ID 和角色信息

### 9.2 Token 的失效处理

- 用户可以主动通过 **POST /token/invalidate** 使 token 失效
- 已失效的 token 将无法继续使用

### 9.3 需要 Token 认证的接口

以下接口需要在请求头中携带有效的 token：

1. **POST /raiseHand** - 举手请求接口

   - 需要学生 token
   - token 中的用户 ID 必须与请求中的学生 ID 匹配

2. **POST /admin/update** - 更新管理员信息接口

   - 需要管理员 token
   - token 中的用户 ID 必须与待更新的管理员 ID 匹配
   - 仅能更新自己的信息

3. **POST /exam/create** - 创建考试接口

   - 需要管理员 token

4. **PUT /exam/streamtype** - 更新考试流类型接口

   - 需要管理员 token

5. **POST /question/create** - 创建题目接口

   - 需要管理员 token

6. **GET /question/list** - 获取考试题目列表接口

   - 需要管理员 token

7. **GET /question/{id}** - 获取题目详情接口

   - 需要管理员 token

8. **PUT /question/{id}** - 更新题目接口

   - 需要管理员 token

9. **DELETE /question/{id}** - 删除题目接口

   - 需要管理员 token

10. **POST /screen/init** - 初始化屏幕流接口

    - 需要学生 token

11. **GET /screen/active** - 获取活跃屏幕流接口

    - 需要管理员 token

12. **POST /screen/stop** - 停止屏幕流接口
    - 需要学生或管理员 token
    - 学生只能停止自己的屏幕流
    - 管理员可以停止任何屏幕流

## 10. 角色说明

- **student**：考生角色，用于进入考试页面。
- **admin**：管理员角色，用于进入监考或管理页面。
- 可通过生成密码时指定 `role`，默认为 "student"；生成管理员密码时需指定为 "admin"。

## 11. 系统初始化说明

- 系统首次启动时，会自动检测数据库中是否存在管理员备份；如果不存在，则采用默认配置（`admin`/`admin123`/`EXAM001`）创建顶级管理员，并备份管理员信息。
- 若管理员使用默认信息登录，验证接口返回的 `is_default` 字段将为 `true`，客户端应提醒管理员尽快通过 **POST /admin/update** 修改默认信息，确保系统安全。
- 测试阶段可通过 **DELETE /admin/deleteAll** 接口清除所有管理员数据，再次启动服务将重新生成默认管理员。
