import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/foundation.dart';

/// The router's view of the module flags: whether the first answer has
/// arrived and, once it has, which modules are on.
///
/// While [isResolved] is false a gated location waits on the splash. The
/// wait is bounded: [startDeadline] resolves to "nothing enabled" after
/// [timeout], so a server that never answers fails closed to not found.
/// Notifies on every change, which re-runs the router's redirect.
class RouterFlags extends ChangeNotifier {
  /// Starts unresolved, or resolved with [initial] when flags are known.
  new({required this.timeout, Set<AppModule>? initial})
    : _enabled = initial ?? const {},
      _isResolved = initial != null;

  /// The longest a gated location waits for the first answer.
  final Duration timeout;

  Set<AppModule> _enabled;
  bool _isResolved;
  Timer? _deadline;

  /// Whether the first flags answer (or the timeout) has arrived.
  bool get isResolved => _isResolved;

  /// Whether [module] is on. Always `false` before [isResolved].
  bool isEnabled(AppModule module) => _enabled.contains(module);

  /// Records a flags answer from the server.
  void resolve(Set<AppModule> enabled) {
    _deadline?.cancel();
    _enabled = enabled;
    _isResolved = true;
    notifyListeners();
  }

  /// Starts the bounded wait once; a no-op when resolved or already waiting.
  /// Started lazily, by the first gated location, so a boot that never
  /// visits one leaves no timer behind.
  void startDeadline() {
    if (_isResolved || _deadline != null) return;
    _deadline = Timer(timeout, () => resolve(const {}));
  }

  @override
  void dispose() {
    _deadline?.cancel();
    super.dispose();
  }
}
