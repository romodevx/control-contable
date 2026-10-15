import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';

/// Housing rent carries no IVA (LIVA 20 fr. II): while a form says the
/// rent is for housing, its IVA rate field is held at exempt; when it stops,
/// the rate goes back to where it was (or to [_fallback] if that was
/// already exempt).
class HousingRentRate {
  /// Holds [_rate], falling back to what [_fallback] says.
  new(this._rate, this._fallback);

  final ChoiceFieldController _rate;
  final IvaRate Function() _fallback;
  IvaRate? _before;

  /// Follows the form's answer: [isHousing] holds the rate at exempt.
  void follow({required bool isHousing}) {
    if (isHousing && _before == null) {
      _before = IvaRate.fromJson(_rate.value ?? '');
      _rate.value = IvaRate.exempt.json;
    } else if (!isHousing && _before != null) {
      final before = _before!;
      _before = null;
      final restored = before == IvaRate.exempt || before == IvaRate.$unknown
          ? _fallback()
          : before;
      _rate.value = restored.json;
    }
  }
}
