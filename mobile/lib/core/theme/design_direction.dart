import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/theme/directions/confianza_sobria.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// A visual direction: the same tokens, designed once per brightness.
///
/// Swapping the look of the whole app means writing another
/// `directions/<name>.dart` and returning it from [designDirectionProvider].
/// No widget changes.
@immutable
class DesignDirection {
  /// Creates a direction from its light and dark token sets.
  const new({required this.name, required this.light, required this.dark});

  /// Human-readable name, as in DIRECTIONS.md.
  final String name;

  /// Tokens for light mode.
  final AppTokens light;

  /// Tokens for dark mode (designed, not inverted).
  final AppTokens dark;
}

/// The active direction. "A · Confianza sobria" until the owner picks
/// another one in PR #3.
final designDirectionProvider = Provider<DesignDirection>(
  (ref) => confianzaSobria,
  name: 'designDirectionProvider',
);
