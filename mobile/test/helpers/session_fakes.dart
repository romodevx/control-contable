import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';

/// A token pair as `login`, `verifyEmail` or `refreshSession` answer it.
TokenPair tokenPair({required String access, required String refresh}) =>
    TokenPair(
      accessToken: access,
      accessExpiresIn: 900,
      refreshToken: refresh,
      refreshExpiresIn: 2592000,
      tokenType: 'Bearer',
    );

/// A [RefreshTokenStore] held in memory, so tests can read what the app
/// kept. A [readError] makes every read fail with it.
class InMemoryRefreshTokenStore implements RefreshTokenStore {
  /// A store that starts with [token] (none by default).
  new([this.token]);

  /// The token kept, if any.
  String? token;

  /// Every token written, in order.
  final List<String> written = [];

  /// Thrown by [read] when set.
  Exception? readError;

  @override
  Future<String?> read() async {
    if (readError case final error?) throw error;
    return token;
  }

  @override
  Future<void> write(String refreshToken) async {
    written.add(refreshToken);
    token = refreshToken;
  }

  @override
  Future<void> delete() async => token = null;
}
