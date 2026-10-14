import 'package:control_contable/api/generated/models/person_type.dart';

/// Length of a persona física RFC (4 letters, date, homoclave, digit).
const int _fisicaLength = 13;

/// Length of a persona moral RFC (3 letters, date, homoclave, digit).
const int _moralLength = 12;

/// What may sit between the characters of a typed RFC: whitespace (also a
/// non-breaking space) and dashes.
final RegExp _separators = RegExp(r'[\s -]');

/// Letters of the name part, the 6-digit date, the 2-char homoclave and the
/// check digit (`[0-9A]`). MODULE-taxpayers §2.2.
final RegExp _fisicaShape = RegExp(r'^[A-ZÑ&]{4}\d{6}[A-Z0-9]{2}[0-9A]$');
final RegExp _moralShape = RegExp(r'^[A-ZÑ&]{3}\d{6}[A-Z0-9]{2}[0-9A]$');

/// SAT generic RFCs (público en general, extranjero): they identify no
/// taxpayer.
const Set<String> _genericRfcs = {'XAXX010101000', 'XEXX010101000'};

/// Length of the `YYMMDD` date part.
const int _dateLength = 6;

const int _monthsPerYear = 12;

/// The century a two-digit year is read in: only its leap cycle matters.
const int _century = 2000;

/// Why a typed RFC cannot be sent. Mirrors the server's `details[].code`
/// (`rfc_length_mismatch`, `rfc_format`, `rfc_date`, `rfc_generic`); the
/// server stays the authority. The check digit is not checked here: a
/// mismatch never rejects (D-RFC), the server warns about it.
enum RfcProblem {
  /// Wrong length for the person type.
  length,

  /// Characters outside the RFC shape.
  format,

  /// The `YYMMDD` part is not a real date.
  date,

  /// A SAT generic RFC.
  generic,
}

/// The RFC length for [personType]: 13 física, 12 moral.
int rfcLength(PersonType personType) =>
    personType == PersonType.moral ? _moralLength : _fisicaLength;

/// [typed] as the server normalizes it: upper-case, without spaces or
/// dashes (the user may type it either way).
String normalizeRfc(String typed) =>
    typed.replaceAll(_separators, '').toUpperCase();

/// The problem a normalized [rfc] has for [personType], or `null` when the
/// client finds none.
RfcProblem? checkRfc(String rfc, PersonType personType) {
  if (rfc.runes.length != rfcLength(personType)) return RfcProblem.length;
  final shape = personType == PersonType.moral ? _moralShape : _fisicaShape;
  if (!shape.hasMatch(rfc)) return RfcProblem.format;
  if (_genericRfcs.contains(rfc)) return RfcProblem.generic;
  final dateStart = rfc.runes.length - _dateLength - 3;
  final date = rfc.substring(dateStart, dateStart + _dateLength);
  return _isRealDate(date) ? null : RfcProblem.date;
}

bool _isRealDate(String yymmdd) {
  final year = int.parse(yymmdd.substring(0, 2));
  final month = int.parse(yymmdd.substring(2, 4));
  final day = int.parse(yymmdd.substring(4));
  if (month < 1 || month > _monthsPerYear || day < 1) return false;
  // Day 0 of the next month is the last day of this one.
  return day <= DateTime.utc(_century + year, month + 1, 0).day;
}
