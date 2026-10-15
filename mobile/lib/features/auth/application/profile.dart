import 'package:control_contable/api/generated/models/me.dart';
import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The signed-in user's profile and access state (`GET /v1/me`), read once
/// per visit to the account screen. A failed load shows a retry button: no
/// automatic retry.
final FutureProvider<Me> profileProvider = FutureProvider.autoDispose<Me>(
  (ref) => ref.watch(profileRepositoryProvider).getMe(),
  name: 'profileProvider',
  retry: _noAutomaticRetry,
);

Duration? _noAutomaticRetry(int retryCount, Object error) => null;

/// The [User] part of [Me]: what the profile form shows and edits.
extension MeProfile on Me {
  /// The profile fields, without the access state.
  User get profile =>
      User(id: id, email: email, name: name, createdAt: createdAt);
}
