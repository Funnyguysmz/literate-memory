import 'package:flutter/material.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _examIdController = TextEditingController();
  final _userIdController = TextEditingController();
  final _passwordController = TextEditingController();

  String? token;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('登入考试')),
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
                        '登入考试',
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
                          labelText: '学号/工号',
                          prefixIcon: Icon(Icons.person),
                          border: OutlineInputBorder(),
                        ),
                        validator: (value) => (value == null || value.isEmpty)
                            ? '请输入学号/工号'
                            : null,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _passwordController,
                        decoration: const InputDecoration(
                          labelText: '密码',
                          prefixIcon: Icon(Icons.lock),
                          border: OutlineInputBorder(),
                        ),
                        obscureText: true,
                        validator: (value) =>
                            (value == null || value.isEmpty) ? '请输入密码' : null,
                      ),
                      const SizedBox(height: 32),
                      SizedBox(
                        width: double.infinity,
                        height: 50,
                        child: ElevatedButton(
                          onPressed: () {
                            if (_formKey.currentState!.validate()) {
                              // 模拟接口返回并保存 token
                              String userId = _userIdController.text;
                              Map<String, dynamic> response = {
                                "message": "验证成功",
                                "role": userId.startsWith("admin")
                                    ? "admin"
                                    : "student",
                                "is_default": userId.startsWith("admin"),
                                "examid": _examIdController.text,
                                "userid": userId,
                                "token": "sample_token_123"
                              };
                              token = response["token"];
                              if (response["role"] == "student") {
                                Navigator.pushReplacementNamed(
                                    context, '/examMain');
                              } else {
                                Navigator.pushReplacementNamed(
                                    context, '/monitor');
                              }
                            }
                          },
                          child: const Text('登入'),
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
    _passwordController.dispose();
    super.dispose();
  }
}
