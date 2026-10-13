import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The current time. Read through [clockProvider] so tests control it.
typedef Clock = DateTime Function();

/// The app's clock (`DateTime.now`); tests override it with a fake.
final clockProvider = Provider<Clock>(
  (ref) => DateTime.now,
  name: 'clockProvider',
);
