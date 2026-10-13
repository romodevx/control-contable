import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Where the session's refresh token survives an app restart. The access
/// token never comes here: it lives in memory only (`Session`).
abstract interface class RefreshTokenStore {
  /// The kept refresh token, or `null` when there is none.
  Future<String?> read();

  /// Keeps [refreshToken], replacing the previous one (rotation).
  Future<void> write(String refreshToken);

  /// Forgets the refresh token.
  Future<void> delete();

  /// Forgets everything the app keeps in this store: the last resort when
  /// [delete] keeps failing, so a refresh token never outlives a sign-out.
  Future<void> wipeAll();
}

/// [RefreshTokenStore] over `flutter_secure_storage`: the Keychain on iOS,
/// Keystore-wrapped keys on Android (ADR-0001; mobile/README.md, "Session
/// tokens"). The only place in `lib/` that touches the platform's secure
/// storage (`test/platform/token_storage_test.dart`); never
/// `shared_preferences`.
class SecureRefreshTokenStore implements RefreshTokenStore {
  /// The store on [storage] (the platform's by default).
  new([FlutterSecureStorage? storage])
    : _storage = storage ?? const FlutterSecureStorage(iOptions: appleOptions);

  /// The Keychain item stays on this device: readable after the first
  /// unlock (so a background refresh works), never synced to iCloud and
  /// never restored onto another device. Android's equivalent are the
  /// backup exclusions in `res/xml` (`android_backup_test.dart`).
  static const IOSOptions appleOptions = IOSOptions(
    accessibility: KeychainAccessibility.first_unlock_this_device,
  );

  static const String _key = 'session.refresh_token';

  final FlutterSecureStorage _storage;

  @override
  Future<String?> read() => _storage.read(key: _key);

  @override
  Future<void> write(String refreshToken) =>
      _storage.write(key: _key, value: refreshToken);

  @override
  Future<void> delete() => _storage.delete(key: _key);

  @override
  Future<void> wipeAll() => _storage.deleteAll();
}

/// The refresh token store. Tests override it with an in-memory store.
final refreshTokenStoreProvider = Provider<RefreshTokenStore>(
  (ref) => SecureRefreshTokenStore(),
  name: 'refreshTokenStoreProvider',
);
