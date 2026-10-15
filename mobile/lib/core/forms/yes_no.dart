/// The wire value of a "Sí" in a yes/no choice field.
const String yes = 'true';

/// The wire value of a "No" in a yes/no choice field.
const String no = 'false';

/// A bool as the wire value of a yes/no choice field.
extension YesNoChoice on bool {
  /// [yes] or [no].
  String get choice => this ? yes : no;
}
