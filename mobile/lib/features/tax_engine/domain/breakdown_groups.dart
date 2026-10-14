import 'package:control_contable/api/generated/models/breakdown_line.dart';
import 'package:control_contable/api/generated/models/breakdown_line_tax.dart';

/// The lines of one tax of a breakdown, in `seq` order.
typedef BreakdownGroup = ({BreakdownLineTax tax, List<BreakdownLine> lines});

/// The order the groups are read in: the regime's ISR, the month's IVA,
/// then the withholdings to remit (paid apart from the taxes).
const List<BreakdownLineTax> _groupOrder = [
  BreakdownLineTax.isr,
  BreakdownLineTax.iva,
  BreakdownLineTax.withholding,
];

/// [lines] split by tax, each group in `seq` order. The server sends them
/// so; sorting (O(n log n)) guards the reading order anyway. A tax with no
/// line has no group; a tax this app version does not know is left out.
List<BreakdownGroup> groupLines(List<BreakdownLine> lines) {
  final sorted = [...lines]..sort((a, b) => a.seq.compareTo(b.seq));
  return [
    for (final tax in _groupOrder)
      if (sorted.where((line) => line.tax == tax).toList() case final group
          when group.isNotEmpty)
        (tax: tax, lines: group),
  ];
}
