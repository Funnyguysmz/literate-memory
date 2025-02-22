import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:oj_client/config/base_config.dart';

class BindingPage extends StatefulWidget {
  const BindingPage({super.key});

  @override
  State<BindingPage> createState() => _BindingPageState();
}

class _BindingPageState extends State<BindingPage> {
  final _formKey = GlobalKey<FormState>();
  final _examIdController = TextEditingController();
  final _userIdController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('绑定考试')),
      body: Center(
        child: SingleChildScrollView(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 400),
            child: Card(
              elevation: 2,
              margin: const EdgeInsets.all(24.0),
              child: Padding(
                padding: const EdgeInsets.all(32.0),
                child: Form(
                  key: _formKey,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        '绑定考试',
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      const SizedBox(height: 32),
                      TextFormField(
                        controller: _examIdController,
                        decoration: const InputDecoration(
                          labelText: '考试场次ID',
                          prefixIcon: Icon(Icons.numbers),
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) => (value == null || value.isEmpty)
                            ? '请输入考试场次ID'
                            : null,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _userIdController,
                        decoration: const InputDecoration(
                          labelText: '学号',
                          prefixIcon: Icon(Icons.person),
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) =>
                            (value == null || value.isEmpty) ? '请输入学号' : null,
                      ),
                      const SizedBox(height: 32),
                      SizedBox(
                        width: double.infinity,
                        height: 50,
                        child: ElevatedButton(
                          onPressed: () async {
                            if (_formKey.currentState!.validate()) {
                              final examid = _examIdController.text;
                              final stuid = _userIdController.text;
                              try {
                                // 调用 POST /exam/bind 接口进行绑定
                                final bindResponse = await http.post(
                                  Uri.parse('${BaseConfig.baseUrl}/exam/bind'),
                                  headers: {"Content-Type": "application/json"},
                                  body: jsonEncode({
                                    "examid": examid,
                                    "stuid": stuid,
                                  }),
                                );
                                if (bindResponse.statusCode != 200) {
                                  final errorMsg =
                                      jsonDecode(bindResponse.body)["error"] ??
                                          "绑定失败";
                                  showDialog(
                                    context: context,
                                    builder: (_) => AlertDialog(
                                      title: const Text('绑定失败'),
                                      content: Text(errorMsg),
                                      actions: [
                                        TextButton(
                                          onPressed: () =>
                                              Navigator.pop(context),
                                          child: const Text('确定'),
                                        )
                                      ],
                                    ),
                                  );
                                  return;
                                }
                                // 绑定成功后调用 GET /password?stuid=xxx 接口获取考试密码
                                final passwordResponse = await http.get(
                                  Uri.parse(
                                      '${BaseConfig.baseUrl}/password?stuid=$stuid'),
                                );
                                if (passwordResponse.statusCode != 200) {
                                  final errorMsg = jsonDecode(
                                          passwordResponse.body)["error"] ??
                                      "获取密码失败";
                                  showDialog(
                                    context: context,
                                    builder: (_) => AlertDialog(
                                      title: const Text('获取密码失败'),
                                      content: Text(errorMsg),
                                      actions: [
                                        TextButton(
                                          onPressed: () =>
                                              Navigator.pop(context),
                                          child: const Text('确定'),
                                        )
                                      ],
                                    ),
                                  );
                                  return;
                                }
                                final examPassword = jsonDecode(
                                    passwordResponse.body)["password"];
                                // 弹出绑定成功对话框并显示登录密码，然后跳转到登入考试页面
                                showDialog(
                                  context: context,
                                  builder: (_) => AlertDialog(
                                    title: const Text('绑定成功'),
                                    content: Text(
                                        '绑定成功！您的登录密码为：$examPassword\n请记住密码'),
                                    actions: [
                                      TextButton(
                                        onPressed: () {
                                          Navigator.pop(context);
                                          Navigator.pushReplacementNamed(
                                              context, '/examLogin');
                                        },
                                        child: const Text('确定'),
                                      ),
                                    ],
                                  ),
                                );
                              } catch (e) {
                                showDialog(
                                  context: context,
                                  builder: (_) => AlertDialog(
                                    title: const Text('请求异常'),
                                    content: Text(e.toString()),
                                    actions: [
                                      TextButton(
                                        onPressed: () => Navigator.pop(context),
                                        child: const Text('确定'),
                                      )
                                    ],
                                  ),
                                );
                              }
                            }
                          },
                          child: const Text('绑定'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  void dispose() {
    _examIdController.dispose();
    _userIdController.dispose();
    super.dispose();
  }
}
