import 'package:control_contable/api/generated/models/breakdown_line.dart';
import 'package:control_contable/api/generated/models/breakdown_line_tax.dart';
import 'package:control_contable/api/generated/models/citation.dart';
import 'package:control_contable/features/tax_engine/domain/breakdown_groups.dart';
import 'package:control_contable/features/tax_engine/domain/legal_citation.dart';
import 'package:control_contable/features/tax_engine/domain/rate_percent.dart';
import 'package:flutter_test/flutter_test.dart';

import 'breakdown_fixtures.dart';

BreakdownLine _line(Map<String, Object?> json) => BreakdownLine.fromJson(json);

void main() {
  group('formatRatePercent', () {
    test('moves the decimal point on the string, never through a float', () {
      expect(formatRatePercent('0.2136'), '21.36 %');
      expect(formatRatePercent('0.16'), '16 %');
      expect(formatRatePercent('0.1'), '10 %');
      expect(formatRatePercent('0.0125'), '1.25 %');
      expect(formatRatePercent('1'), '100 %');
      expect(formatRatePercent('0'), '0 %');
      expect(formatRatePercent('0.333333'), '33.3333 %');
    });
  });

  group('citations', () {
    const article106 = Citation(
      law: 'LISR',
      article: '106',
      note: 'párrafo quinto',
    );
    const annex = Citation(
      law: 'Anexo 8 RMF 2026',
      article: 'B.VI',
      note: null,
    );

    test('short form names the law and the article', () {
      expect(shortCitation(article106), 'LISR 106');
      expect(shortCitation(annex), 'Anexo 8 RMF 2026 B.VI');
    });

    test('long form spells a known law out, with its note', () {
      expect(
        fullCitation(article106),
        'Ley del Impuesto sobre la Renta, artículo 106, párrafo quinto',
      );
      expect(
        fullCitation(const Citation(law: 'LIVA', article: '5-D', note: null)),
        'Ley del Impuesto al Valor Agregado, artículo 5-D',
      );
      expect(fullCitation(annex), 'Anexo 8 RMF 2026, B.VI');
    });

    test('known laws link to their official text; others do not', () {
      expect(
        officialLaw('LISR')?.url,
        Uri.parse('https://www.diputados.gob.mx/LeyesBiblio/pdf/LISR.pdf'),
      );
      expect(
        officialLaw('RLIVA')?.url,
        Uri.parse(
          'https://www.diputados.gob.mx/LeyesBiblio/regley/Reg_LIVA.pdf',
        ),
      );
      expect(officialLaw('Anexo 8 RMF 2026'), isNull);
    });
  });

  group('groupLines', () {
    test('splits by tax and orders each group by seq', () {
      final groups = groupLines([
        _line(withholdingLines.single),
        _line(ivaLines612().last),
        _line(isrLines612[2]),
        _line(isrLines612[0]),
        _line(ivaLines612().first),
      ]);

      expect(groups.map((g) => g.tax), [
        BreakdownLineTax.isr,
        BreakdownLineTax.iva,
        BreakdownLineTax.withholding,
      ]);
      expect(groups[0].lines.map((l) => l.seq), [1, 3]);
      expect(groups[1].lines.map((l) => l.seq), [10, 12]);
      expect(groups[2].lines.map((l) => l.seq), [13]);
    });

    test('leaves out a tax with no lines', () {
      final groups = groupLines([_line(isrLines612.first)]);

      expect(groups.single.tax, BreakdownLineTax.isr);
    });
  });
}
