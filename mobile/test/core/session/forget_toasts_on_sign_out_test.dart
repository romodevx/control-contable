import 'package:control_contable/core/session/forget_toasts_on_sign_out.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_error_reporter.dart';
import '../../helpers/session_fakes.dart';

/// Tech-lead gate M3-12/M4-13 (completes QA M3/M4 D4): the Undo toast of a
/// record list never auto-dismisses and lives on the app's root messenger,
/// so it outlived sign-out: the next account on the device saw "2 ingresos
/// eliminados · Deshacer", and *Deshacer* restored with its token. Signing
/// out now closes every toast; a token refresh keeps them.
void main() {
  late Session session;
  late ScaffoldMessengerState messenger;

  Future<void> pump(WidgetTester tester) async {
    session = Session(
      store: InMemoryRefreshTokenStore(),
      rotate: (_) async => tokenPair(access: 'a', refresh: 'r'),
      reporter: RecordingErrorReporter(),
    );
    addTearDown(session.dispose);
    await session.start(tokenPair(access: 'a1', refresh: 'r1'));
    await tester.pumpWidget(
      MaterialApp(
        builder: (context, child) =>
            ForgetToastsOnSignOut(session: session, child: child!),
        home: Scaffold(
          body: Builder(
            builder: (context) {
              messenger = ScaffoldMessenger.of(context);
              return const SizedBox.shrink();
            },
          ),
        ),
      ),
    );
    messenger.showSnackBar(
      const SnackBar(content: Text('2 ingresos eliminados'), persist: true),
    );
    await tester.pumpAndSettle();
  }

  testWidgets('signing out closes a pending toast', (tester) async {
    await pump(tester);
    expect(find.text('2 ingresos eliminados'), findsOneWidget);

    await session.end();
    await tester.pumpAndSettle();

    expect(find.text('2 ingresos eliminados'), findsNothing);
  });

  testWidgets('a token refresh keeps it', (tester) async {
    await pump(tester);

    await session.refresh();
    await tester.pumpAndSettle();

    expect(find.text('2 ingresos eliminados'), findsOneWidget);
  });
}
