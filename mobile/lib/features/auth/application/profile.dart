import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The signed-in user's profile, read once per visit to the account screen.
/// A failed load shows a retry button: no automatic retry.
final FutureProvider<User> profileProvider = FutureProvider.autoDispose<User>(
  (ref) => ref.watch(profileRepositoryProvider).getMe(),
  name: 'profileProvider',
  retry: _noAutomaticRetry,
);

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
