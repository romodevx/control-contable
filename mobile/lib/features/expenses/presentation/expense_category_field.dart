import 'dart:async';

import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Share of the screen the picker sheet takes.
const double _sheetHeightFactor = 0.85;

/// The category of an expense: its label, a 44 px button naming the choice
/// that opens a searchable picker ([categories], most used first), and the
/// error under it (also in the form's error summary).
class ExpenseCategoryField extends StatelessWidget {
  /// The picker for [field] over [categories].
  const new({required this.field, required this.categories, super.key});

  /// The chosen code and its error.
  final ChoiceFieldController field;

  /// What can be chosen, in the order the picker lists them.
  final List<ExpenseCategory> categories;

  Future<void> _pick(BuildContext context) async {
    final code = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (_) =>
          _CategoryPicker(categories: categories, selected: field.value),
    );
    if (code != null) field.value = code;
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: field,
    builder: (context, _) {
      final theme = Theme.of(context);
      final l10n = AppLocalizations.of(context);
      final colors = context.tokens.colors;
      final error = field.error;
      final chosen = categories
          .where((c) => c.code == field.value)
          .firstOrNull
          ?.name;
      final shown = chosen ?? l10n.expenseCategoryNone;
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          ExcludeSemantics(
            child: Text(
              field.label,
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          const SizedBox(height: AppSpacing.xxs),
          Semantics(
            button: true,
            label: '${field.label}: $shown',
            excludeSemantics: true,
            child: OutlinedButton(
              focusNode: field.focusNode,
              style: OutlinedButton.styleFrom(
                alignment: Alignment.centerLeft,
                minimumSize: const Size.fromHeight(AppSizes.minTarget),
                side: error == null ? null : BorderSide(color: colors.error),
              ),
              onPressed: () => unawaited(_pick(context)),
              child: Row(
                children: [
                  Expanded(child: Text(shown)),
                  const Icon(Icons.expand_more),
                ],
              ),
            ),
          ),
          if (error != null)
            Semantics(
              liveRegion: true,
              child: Padding(
                padding: const EdgeInsets.only(top: AppSpacing.xxs),
                child: Text(
                  describeFieldError(error, l10n),
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: colors.error,
                  ),
                ),
              ),
            ),
        ],
      );
    },
  );
}

/// The sheet: a title, a search field and the categories that match it;
/// choosing one closes the sheet with its code.
class _CategoryPicker extends StatefulWidget {
  const new({required this.categories, required this.selected});

  final List<ExpenseCategory> categories;
  final String? selected;

  @override
  State<_CategoryPicker> createState() => _CategoryPickerState();
}

class _CategoryPickerState extends State<_CategoryPicker> {
  String _query = '';

  List<ExpenseCategory> get _matches {
    final query = foldForSearch(_query.trim());
    if (query.isEmpty) return widget.categories;
    return [
      for (final category in widget.categories)
        if (foldForSearch(category.name).contains(query)) category,
    ];
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final media = MediaQuery.of(context);
    final matches = _matches;
    return Padding(
      padding: EdgeInsets.only(bottom: media.viewInsets.bottom),
      child: SizedBox(
        height: media.size.height * _sheetHeightFactor,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.all(AppSpacing.gutter),
              child: Text(
                l10n.expenseCategoryPickerTitle,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.gutter,
              ),
              child: TextField(
                keyboardType: TextInputType.text,
                textInputAction: TextInputAction.search,
                decoration: InputDecoration(
                  labelText: l10n.expenseCategorySearch,
                  prefixIcon: const Icon(Icons.search),
                ),
                onChanged: (value) => setState(() => _query = value),
              ),
            ),
            const SizedBox(height: AppSpacing.xs),
            if (matches.isEmpty)
              Padding(
                padding: const EdgeInsets.all(AppSpacing.gutter),
                child: Text(l10n.expenseCategoryNoMatch),
              ),
            Expanded(
              child: ListView(
                children: [
                  for (final category in matches)
                    ListTile(
                      title: Text(category.name),
                      selected: category.code == widget.selected,
                      trailing: category.code == widget.selected
                          ? const Icon(Icons.check)
                          : null,
                      onTap: () => Navigator.of(context).pop(category.code),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

const Map<String, String> _accents = {
  'á': 'a',
  'é': 'e',
  'í': 'i',
  'ó': 'o',
  'ú': 'u',
  'ü': 'u',
};

/// [text] lowercased without Spanish accents, so "papeleria" finds
/// "Papelería".
String foldForSearch(String text) =>
    [for (final char in text.toLowerCase().split('')) _accents[char] ?? char]
        .join();
