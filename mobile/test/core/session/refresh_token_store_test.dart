import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  setUp(() => FlutterSecureStorage.setMockInitialValues({}));

  test('writes, reads back and deletes the refresh token', () async {
    final store = SecureRefreshTokenStore();

    expect(await store.read(), isNull);
    await store.write('refresh-1');
    expect(await store.read(), 'refresh-1');
    await store.delete();
    expect(await store.read(), isNull);
  });

  test('a rotation replaces the stored token', () async {
    final store = SecureRefreshTokenStore();

    await store.write('refresh-1');
    await store.write('refresh-2');

    expect(await store.read(), 'refresh-2');
  });

  test('the Keychain item never syncs nor migrates to another device', () {
    const options = SecureRefreshTokenStore.appleOptions;

    expect(
      options.accessibility,
      KeychainAccessibility.first_unlock_this_device,
    );
    expect(options.synchronizable, isFalse);
  });
}
