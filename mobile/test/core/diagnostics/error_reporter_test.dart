import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('a report names the operation and the error type only', () {
    const token = 'refresh-secret';
    final error = PlatformException(code: 'read', message: token);

    final line = describeAbsorbedError('signOut.revoke', error);

    expect(line, 'signOut.revoke failed: PlatformException');
    expect(line, isNot(contains(token)));
  });
}
