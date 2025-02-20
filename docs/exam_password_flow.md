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

## 5. 异常情况处理

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

## 6. 管理员相关接口

### 6.1 顶级管理员信息查看

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

### 6.2 创建新考试

- **POST /exam/create**

```json
请求示例：
{
    "examid": "EXAM002"
}

返回示例：
{
    "message": "考试创建成功",
    "examid": "EXAM002"
}
```

### 6.3 学生绑定考试

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

### 6.4 更新管理员信息

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

### 6.5 删除管理员数据（测试接口）

- 用于开发测试阶段，**DELETE /admin/deleteAll** 可清除所有管理员数据。

```json
返回示例：
{
    "message": "所有管理员数据已删除"
}
```

## 7. 角色说明

- **student**：考生角色，用于进入考试页面。
- **admin**：管理员角色，用于进入监考或管理页面。
- 可通过生成密码时指定 `role`，默认为 "student"；生成管理员密码时需指定为 "admin"。

## 8. 系统初始化说明

- 系统首次启动时，会自动检测数据库中是否存在管理员备份；如果不存在，则采用默认配置（`admin`/`admin123`/`EXAM001`）创建顶级管理员，并备份管理员信息。
- 若管理员使用默认信息登录，验证接口返回的 `is_default` 字段将为 `true`，客户端应提醒管理员尽快通过 **POST /admin/update** 修改默认信息，确保系统安全。
- 测试阶段可通过 **DELETE /admin/deleteAll** 接口清除所有管理员数据，再次启动服务将重新生成默认管理员。
