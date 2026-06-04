import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:chatbot_admin_mobile/auth/auth_store.dart';
import 'package:chatbot_admin_mobile/main.dart';

void main() {
  testWidgets('启动后展示登录或首页', (WidgetTester tester) async {
    TestWidgetsFlutterBinding.ensureInitialized();
    SharedPreferences.setMockInitialValues({});
    final store = AuthStore();
    await store.load();
    await tester.pumpWidget(ChatbotAdminApp(authStore: store));
    await tester.pump();
    expect(find.textContaining('登录'), findsWidgets);
  });
}
