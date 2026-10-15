import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/widgets/polite_announcement.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

/// The key tests and the error summary find the card by.
const ValueKey<String> withholdingCardKey = ValueKey('withholding-card');

/// What the card says, already in es-MX: the forms build it from the
/// server's derivation (ADR-0008), never from a rate table of their own.
@immutable
class WithholdingCardCopy {
  /// A card reading [headline], then [details], [explanations], [warnings]
  /// and a [notice] about the figures' state.
  const new({
    this.headline = const [],
    this.details = const [],
    this.explanations = const [],
    this.warnings = const [],
    this.notice,
  });

  /// The amounts ("Te retienen…", "Recibes…"): announced when they change.
  final List<String> headline;

  /// Further amounts or dates ("Entregas al SAT…").
  final List<String> details;

  /// One line per rule applied, as the server words it.
  final List<String> explanations;

  /// Non-blocking ⚠️ lines: announced as they appear.
  final List<String> warnings;

  /// Why there are no (fresh) figures: incomplete facts, offline.
  final String? notice;
}

/// The read-only card of derived withholdings (MODULE-withholdings §2.3):
/// no field in it is editable. Its amounts and new warnings are announced
/// politely; a help toggle explains that the law's figure is the one used.
class DerivedWithholdingCard extends StatefulWidget {
  /// The card reading [copy]; [helpLabel] opens [helpText].
  const new({
    required this.copy,
    required this.helpLabel,
    required this.helpText,
    super.key = withholdingCardKey,
  });

  /// What it says.
  final WithholdingCardCopy copy;

  /// "¿Tu factura dice otra cantidad?".
  final String helpLabel;

  /// Why the app's figure stands.
  final String helpText;

  @override
  State<DerivedWithholdingCard> createState() => _DerivedWithholdingCardState();
}

class _DerivedWithholdingCardState extends State<DerivedWithholdingCard> {
  bool _isHelpOpen = false;

  /// The amounts as one sentence, so a screen reader hears them together.
  static List<String> _spoken(WithholdingCardCopy copy) =>
      copy.headline.isEmpty ? const [] : [copy.headline.join('. ')];

  @override
  void initState() {
    super.initState();
    announceNewPolitely(this, widget.copy.warnings);
  }

  @override
  void didUpdateWidget(DerivedWithholdingCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    final before = oldWidget.copy;
    final now = widget.copy;
    if (!listEquals(now.headline, before.headline)) {
      announceNewPolitely(this, _spoken(now));
    }
    announceNewPolitely(this, now.warnings, before: before.warnings);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final tokens = context.tokens;
    final colors = tokens.colors;
    final copy = widget.copy;
    final strong = theme.textTheme.bodyLarge?.copyWith(
      fontWeight: FontWeight.w600,
    );
    final quiet = theme.textTheme.bodyMedium?.copyWith(color: colors.muted);
    return Semantics(
      container: true,
      liveRegion: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.surface,
          border: Border.all(color: colors.border),
          borderRadius: BorderRadius.circular(tokens.radii.card),
        ),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              for (final line in copy.headline) Text(line, style: strong),
              for (final line in copy.details)
                Text(line, style: theme.textTheme.bodyMedium),
              if (copy.notice case final notice?)
                Text(notice, style: theme.textTheme.bodyMedium),
              if (copy.explanations.isNotEmpty)
                const SizedBox(height: AppSpacing.xs),
              for (final line in copy.explanations) Text(line, style: quiet),
              for (final warning in copy.warnings) _Warning(text: warning),
              _HelpToggle(
                label: widget.helpLabel,
                isOpen: _isHelpOpen,
                onToggle: () => setState(() => _isHelpOpen = !_isHelpOpen),
              ),
              if (_isHelpOpen) Text(widget.helpText, style: quiet),
            ],
          ),
        ),
      ),
    );
  }
}

/// A ⚠️ line: an icon and words, never color alone.
class _Warning extends StatelessWidget {
  const new({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: AppSpacing.xs),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const ExcludeSemantics(
          child: Icon(Icons.warning_amber_outlined, size: 20),
        ),
        const SizedBox(width: AppSpacing.xs),
        Expanded(
          child: Text(text, style: Theme.of(context).textTheme.bodyMedium),
        ),
      ],
    ),
  );
}

/// The 44 px help button, heard as expanded or collapsed.
class _HelpToggle extends StatelessWidget {
  const new({
    required this.label,
    required this.isOpen,
    required this.onToggle,
  });

  final String label;
  final bool isOpen;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) => Align(
    alignment: Alignment.centerLeft,
    child: MergeSemantics(
      child: Semantics(
        expanded: isOpen,
        child: TextButton(
          style: TextButton.styleFrom(
            minimumSize: const Size(0, AppSizes.minTarget),
          ),
          onPressed: onToggle,
          child: Text(label),
        ),
      ),
    ),
  );
}
