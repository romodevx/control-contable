import 'package:control_contable/api/generated/models/citation.dart';

/// A law the app can name in full and link to its official text.
typedef OfficialLaw = ({String name, Uri url});

/// Where the Chamber of Deputies publishes the consolidated federal laws.
const String _lawsBase = 'https://www.diputados.gob.mx/LeyesBiblio';

/// The laws breakdown lines cite, by the abbreviation the engine uses.
/// Proper names, not copy: they are never translated. A citation of
/// anything else (the RMF annexes, published yearly in the DOF) is shown
/// as given, without a link.
final Map<String, OfficialLaw> _officialLaws = {
  'LISR': (
    name: 'Ley del Impuesto sobre la Renta',
    url: Uri.parse('$_lawsBase/pdf/LISR.pdf'),
  ),
  'LIVA': (
    name: 'Ley del Impuesto al Valor Agregado',
    url: Uri.parse('$_lawsBase/pdf/LIVA.pdf'),
  ),
  'CFF': (
    name: 'Código Fiscal de la Federación',
    url: Uri.parse('$_lawsBase/pdf/CFF.pdf'),
  ),
  'RLISR': (
    name: 'Reglamento de la Ley del Impuesto sobre la Renta',
    url: Uri.parse('$_lawsBase/regley/Reg_LISR.pdf'),
  ),
  'RLIVA': (
    name: 'Reglamento de la Ley del Impuesto al Valor Agregado',
    url: Uri.parse('$_lawsBase/regley/Reg_LIVA.pdf'),
  ),
};

/// The law abbreviated [law] (`LISR`), or `null` when the app does not
/// know it.
OfficialLaw? officialLaw(String law) => _officialLaws[law];

/// [citation] as a short reference: `LISR 106`.
String shortCitation(Citation citation) =>
    '${citation.law} ${citation.article}';

/// [citation] spelled out: `Ley del Impuesto sobre la Renta, artículo 106,
/// párrafo quinto`; an unknown source keeps its own words
/// (`Anexo 8 RMF 2026, B.VI`).
String fullCitation(Citation citation) {
  final law = officialLaw(citation.law);
  final note = citation.note;
  return [
    if (law != null) ...[
      law.name,
      'artículo ${citation.article}',
    ] else ...[
      citation.law,
      citation.article,
    ],
    if (note != null && note.isNotEmpty) note,
  ].join(', ');
}
