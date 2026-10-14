import 'dart:async';

import 'package:control_contable/api/generated/models/citation.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/tax_engine/domain/legal_citation.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// A line's legal basis: collapsed to its short references ("Fundamento:
/// LISR 103, LISR 105"), it opens to each article spelled out with its
/// note and a link to the official text of every known law cited. Screen
/// readers hear it as expanded or collapsed.
class CitationDisclosure extends ConsumerStatefulWidget {
  /// The disclosure of [citations] (at least one).
  const new({required this.citations, super.key});

  /// The line's legal basis, in the server's order.
  final List<Citation> citations;

  @override
  ConsumerState<CitationDisclosure> createState() => _CitationDisclosureState();
}

class _CitationDisclosureState extends ConsumerState<CitationDisclosure> {
  bool _isOpen = false;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final citations = widget.citations;
    final references = citations.map(shortCitation).join(', ');
    // One link per law, however many of its articles the line cites.
    final laws = {for (final citation in citations) ?officialLaw(citation.law)};
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // One node: the button announces its own expanded state.
        MergeSemantics(
          child: Semantics(
            expanded: _isOpen,
            child: TextButton.icon(
              style: TextButton.styleFrom(
                alignment: AlignmentDirectional.centerStart,
                minimumSize: const Size.fromHeight(AppSizes.minTarget),
              ),
              icon: Icon(_isOpen ? Icons.expand_less : Icons.expand_more),
              label: Text(l10n.breakdownLegalBasis(references)),
              onPressed: () => setState(() => _isOpen = !_isOpen),
            ),
          ),
        ),
        if (_isOpen) ...[
          for (final citation in citations)
            Padding(
              padding: const EdgeInsets.only(bottom: AppSpacing.xxs),
              child: Text(fullCitation(citation), style: text.bodyMedium),
            ),
          for (final law in laws)
            Align(
              alignment: AlignmentDirectional.centerStart,
              child: TextButton.icon(
                icon: const Icon(Icons.open_in_new),
                label: Text(l10n.breakdownReadLaw(law.name)),
                onPressed: () =>
                    unawaited(ref.read(urlOpenerProvider)(law.url)),
              ),
            ),
        ],
      ],
    );
  }
}
