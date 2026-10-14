/// The character the server masks the RFC date with.
const String _serverMask = '*';

/// The character the app shows instead: a bullet reads as "hidden" and is
/// not confused with a typed character.
const String _displayMask = '•';

/// How [rfcMasked] (`GOMA******AB6`, from the server) reads on screen:
/// `GOMA••••••AB6`. The full RFC never reaches the app (S-3).
String displayRfc(String rfcMasked) =>
    rfcMasked.replaceAll(_serverMask, _displayMask);
