# 屏幕获取与直播实现方案

本文档详细介绍考试监管系统中屏幕获取与直播功能的简化实现方案，适合个人开发的本科毕设。

## 1. 概述

本系统采用基于截图的屏幕监控方案，而非复杂的视频流直播。这种方案具有以下优势：

- **开发难度低**：无需掌握复杂的音视频编解码知识
- **资源占用少**：对客户端机器性能要求低
- **带宽友好**：间隔发送截图，而非连续视频流
- **实现简单**：基于现有 Flutter 插件即可实现

## 2. 技术选择

### 2.1 屏幕截图技术

- **desktop_screenshot**：Flutter 插件，支持 Windows/macOS/Linux 平台截屏
- **截图频率**：可配置，默认每 5 秒一次（可根据网络条件和监控需求调整）
- **图像处理**：使用`image`包进行图像压缩和调整大小

```dart
import 'package:desktop_screenshot/desktop_screenshot.dart';
import 'package:image/image.dart' as img;

// 屏幕截图示例代码
Future<Uint8List> captureAndCompressScreen() async {
  // 截取屏幕
  final screenshot = await DesktopScreenshot.takeScreenshot();

  // 解码图像
  img.Image? image = img.decodeImage(screenshot);
  if (image == null) return Uint8List(0);

  // 调整大小（降低分辨率）
  img.Image resized = img.copyResize(
    image,
    width: image.width ~/ 2,  // 将宽度缩小一半
    height: image.height ~/ 2 // 将高度缩小一半
  );

  // 压缩为JPEG并返回
  return Uint8List.fromList(img.encodeJpg(resized, quality: 70));
}
```

### 2.2 数据传输

- **WebSocket**：实现客户端和服务端之间的实时通信
- **Base64 编码**：将图像数据编码为文本格式，方便 WebSocket 传输
- **JSON 格式**：封装图像数据和元数据

```dart
import 'dart:convert';

// 发送截图示例代码
void sendScreenshot(WebSocketChannel channel, Uint8List imageBytes) {
  // Base64编码
  String base64Image = base64Encode(imageBytes);

  // 构建JSON消息
  Map<String, dynamic> message = {
    'type': 'screenshot',
    'examId': currentExamId,
    'studentId': studentId,
    'timestamp': DateTime.now().toIso8601String(),
    'imageData': base64Image,
  };

  // 发送消息
  channel.sink.add(jsonEncode(message));
}
```

## 3. 服务端实现

### 3.1 WebSocket 处理器

```go
// 处理客户端WebSocket连接
func handleScreenshotWebSocket(c *gin.Context) {
    // 升级HTTP连接为WebSocket
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    // 处理消息
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }

        // 解析消息
        var data map[string]interface{}
        if err := json.Unmarshal(message, &data); err != nil {
            continue
        }

        // 检查消息类型
        if data["type"] == "screenshot" {
            // 处理屏幕截图
            examId := data["examId"].(string)
            studentId := data["studentId"].(string)
            timestamp := data["timestamp"].(string)
            imageData := data["imageData"].(string)

            // 保存截图或转发给管理员
            saveOrForwardScreenshot(examId, studentId, timestamp, imageData)
        }
    }
}
```

### 3.2 截图存储与转发

```go
// 保存截图或转发给管理员
func saveOrForwardScreenshot(examId, studentId, timestamp, imageData string) {
    // 解码Base64图像
    imageBytes, err := base64.StdEncoding.DecodeString(imageData)
    if err != nil {
        log.Printf("解码图像失败: %v", err)
        return
    }

    // 选项1: 保存到文件系统
    filename := fmt.Sprintf("screenshots/%s/%s/%s.jpg", examId, studentId, timestamp)
    saveImageToFile(filename, imageBytes)

    // 选项2: 转发给在线的管理员
    forwardToAdmins(examId, studentId, timestamp, imageData)
}
```

## 4. 管理员界面实现

### 4.1 监控界面

管理员界面会展示所有考生的截图，并支持以下功能：

- 按考生查看当前截图
- 查看截图历史
- 放大查看特定截图
- 导出截图

```dart
// 管理端接收截图示例
void listenForScreenshots(WebSocketChannel channel) {
  channel.stream.listen((message) {
    Map<String, dynamic> data = jsonDecode(message);

    if (data['type'] == 'screenshot') {
      String studentId = data['studentId'];
      String imageData = data['imageData'];

      // 解码图像并显示
      Uint8List imageBytes = base64Decode(imageData);
      studentScreenshots[studentId] = imageBytes;

      // 触发UI更新
      setState(() {});
    }
  });
}
```

### 4.2 截图显示组件

```dart
// 屏幕截图展示组件
class ScreenshotTile extends StatelessWidget {
  final String studentId;
  final Uint8List imageData;

  const ScreenshotTile({
    Key? key,
    required this.studentId,
    required this.imageData
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          ListTile(
            title: Text('学生: $studentId'),
            subtitle: Text('最近更新: ${DateTime.now().toString()}'),
          ),
          Image.memory(
            imageData,
            fit: BoxFit.cover,
          ),
          ButtonBar(
            children: [
              TextButton(
                onPressed: () {
                  // 放大查看
                },
                child: Text('放大'),
              ),
              TextButton(
                onPressed: () {
                  // 查看历史
                },
                child: Text('历史'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
```

## 5. 系统配置

### 5.1 客户端配置项

- **截图间隔**：控制截图频率
- **图像质量**：控制 JPEG 压缩质量
- **图像分辨率**：控制发送图像的大小

### 5.2 服务端配置项

- **存储策略**：是否保存所有截图，或只保留最近的 N 张
- **转发策略**：是否实时转发给所有管理员，或只按需发送

## 6. 性能优化

### 6.1 带宽优化

- 根据网络状况自动调整截图质量和频率
- 只在检测到屏幕变化时才发送新截图
- 使用差异压缩算法，仅发送与上一张截图的差异部分

### 6.2 存储优化

- 定期清理历史截图
- 对长期保存的截图进行进一步压缩

## 7. 隐私与安全

- 截图传输过程使用 WebSocket 加密通道
- 截图存储时进行加密
- 只有授权管理员可以查看截图
- 考试结束后自动删除所有截图数据

## 8. 开发计划

1. **第一阶段**：实现基本的定时截图和 WebSocket 传输
2. **第二阶段**：添加管理员查看界面
3. **第三阶段**：实现截图历史查看功能
4. **第四阶段**：优化性能和用户体验

这种简化的实现方案既能满足考试监控的基本需求，又大大降低了开发难度，适合个人开发的本科毕设。
