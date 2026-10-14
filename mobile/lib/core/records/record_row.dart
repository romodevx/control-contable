import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

/// One record in a month list, in the columns of its form: the date, what
/// it was, the amount, then its regime (and a status [badge], if any). The
/// row reads as one sentence and opens the record when tapped; its delete
/// button is a separate 44 px target.
///
/// Deleting is reversible (soft delete + Undo), so [onDelete] acts at once:
/// no confirmation here.
class RecordRow extends StatelessWidget {
  /// A row for one record.
  const new({
    required this.date,
    required this.title,
    required this.amountCents,
    required this.tag,
    required this.deleteTooltip,
    required this.onDelete,
    required this.openLabel,
    required this.onOpen,
    this.badge,
    super.key,
  });

  /// The booking date as the user reads it ("5 oct").
  final String date;

  /// The description, or who paid when there is none.
  final String title;

  /// The amount shown (an income's total with IVA).
  final int amountCents;

  /// The regime the record belongs to.
  final String tag;

  /// A status such as "Por cobrar"; words, never color alone.
  final String? badge;

  /// The delete button's tooltip and label (names the record).
  final String deleteTooltip;

  /// Deletes the record.
  final VoidCallback onDelete;

  /// What tapping the row does, for screen readers ("Editar ingreso: …").
  final String openLabel;

  /// Opens the record (its edit page).
  final VoidCallback onOpen;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = context.tokens.colors;
    final badge = this.badge;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: colors.border)),
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.gutter,
          AppSpacing.sm,
          AppSpacing.xxs,
          AppSpacing.sm,
        ),
        child: Row(
          children: [
            Expanded(
              child: Semantics(
                button: true,
                label: openLabel,
                child: InkWell(
                  onTap: onOpen,
                  child: MergeSemantics(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(date, style: theme.textTheme.bodySmall),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(
                              child: Text(
                                title,
                                style: theme.textTheme.bodyLarge,
                              ),
                            ),
                            const SizedBox(width: AppSpacing.xs),
                            Text(
                              formatPesos(amountCents),
                              style: theme.textTheme.bodyLarge,
                            ),
                          ],
                        ),
                        Wrap(
                          spacing: AppSpacing.xs,
                          crossAxisAlignment: WrapCrossAlignment.center,
                          children: [
                            Text(tag, style: theme.textTheme.bodySmall),
                            if (badge != null) _Badge(badge),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
            IconButton(
              tooltip: deleteTooltip,
              icon: const Icon(Icons.delete_outline),
              onPressed: onDelete,
            ),
          ],
        ),
      ),
    );
  }
}

/// A status word in an outlined pill (the outline is ≥ 3:1; the words
/// carry the meaning).
class _Badge extends StatelessWidget {
  const new(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: tokens.colors.outline),
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.xs),
        child: Text(text, style: Theme.of(context).textTheme.labelSmall),
      ),
    );
  }
}
