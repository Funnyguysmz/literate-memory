import 'package:flutter/material.dart';
import 'package:window_size/window_size.dart' as window_size;
import 'package:flutter/foundation.dart';
import 'dart:io';
import 'pages/splash_page.dart';
import 'pages/selector_page.dart';
import 'pages/binding_page.dart';
import 'pages/login_page.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();

  // 设置桌面窗口大小
  if (!kIsWeb && (Platform.isWindows || Platform.isLinux || Platform.isMacOS)) {
    window_size.setWindowTitle('OJ在线考试平台');
    window_size.setWindowMinSize(const Size(800, 600));
    // 设置初始窗口大小（可选）
    window_size.setWindowFrame(const Rect.fromLTWH(0, 0, 1024, 768));
  }

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'OJ在线考试平台',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const SplashPage(),
      routes: {
        '/selector': (context) => const SelectorPage(),
        '/binding': (context) => const BindingPage(),
        '/examLogin': (context) => const LoginPage(),
        // 以下为占位页面，后续实现具体页面
        '/examMain': (context) =>
            const Scaffold(body: Center(child: Text('考试主页面'))),
        '/monitor': (context) =>
            const Scaffold(body: Center(child: Text('监考页面'))),
      },
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text(
              'You have pushed the button this many times:',
            ),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ),
    );
  }
}
