// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_defaults.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomeDefaults {

@JsonKey(name: 'iva_rates_allowed') List<IvaRate> get ivaRatesAllowed;@JsonKey(name: 'suggested_iva_rate') IvaRate get suggestedIvaRate;@JsonKey(name: 'isr_withholding_allowed') bool get isrWithholdingAllowed;/// Fraction of `amount_cents`; `null` when not allowed or when there is no suggestion.
@JsonKey(includeIfNull: true, name: 'isr_withholding_rate') DecimalRate? get isrWithholdingRate;@JsonKey(name: 'iva_withholding_allowed') bool get ivaWithholdingAllowed;/// Fraction of `iva_cents` (2/3 = `0.666667`); `null` when not allowed or when there is no suggestion.
@JsonKey(includeIfNull: true, name: 'iva_withholding_rate') DecimalRate? get ivaWithholdingRate;/// The 625 activity the ISR rate is for; `null` outside 625 or when it is ambiguous.
@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? get platformActivity;
/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeDefaultsCopyWith<IncomeDefaults> get copyWith => _$IncomeDefaultsCopyWithImpl<IncomeDefaults>(this as IncomeDefaults, _$identity);

  /// Serializes this IncomeDefaults to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomeDefaults;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeDefaults&&const DeepCollectionEquality().equals(other.ivaRatesAllowed, _this.ivaRatesAllowed)&&(identical(other.suggestedIvaRate, _this.suggestedIvaRate) || other.suggestedIvaRate == _this.suggestedIvaRate)&&(identical(other.isrWithholdingAllowed, _this.isrWithholdingAllowed) || other.isrWithholdingAllowed == _this.isrWithholdingAllowed)&&(identical(other.isrWithholdingRate, _this.isrWithholdingRate) || other.isrWithholdingRate == _this.isrWithholdingRate)&&(identical(other.ivaWithholdingAllowed, _this.ivaWithholdingAllowed) || other.ivaWithholdingAllowed == _this.ivaWithholdingAllowed)&&(identical(other.ivaWithholdingRate, _this.ivaWithholdingRate) || other.ivaWithholdingRate == _this.ivaWithholdingRate)&&(identical(other.platformActivity, _this.platformActivity) || other.platformActivity == _this.platformActivity));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeDefaults;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.ivaRatesAllowed),_this.suggestedIvaRate,_this.isrWithholdingAllowed,_this.isrWithholdingRate,_this.ivaWithholdingAllowed,_this.ivaWithholdingRate,_this.platformActivity);
}

@override
String toString() {
  final _this = this as IncomeDefaults;
  return 'IncomeDefaults(ivaRatesAllowed: ${_this.ivaRatesAllowed}, suggestedIvaRate: ${_this.suggestedIvaRate}, isrWithholdingAllowed: ${_this.isrWithholdingAllowed}, isrWithholdingRate: ${_this.isrWithholdingRate}, ivaWithholdingAllowed: ${_this.ivaWithholdingAllowed}, ivaWithholdingRate: ${_this.ivaWithholdingRate}, platformActivity: ${_this.platformActivity})';
}


}

/// @nodoc
abstract mixin class $IncomeDefaultsCopyWith<$Res>  {
  factory $IncomeDefaultsCopyWith(IncomeDefaults value, $Res Function(IncomeDefaults) _then) = _$IncomeDefaultsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'iva_rates_allowed') List<IvaRate> ivaRatesAllowed,@JsonKey(name: 'suggested_iva_rate') IvaRate suggestedIvaRate,@JsonKey(name: 'isr_withholding_allowed') bool isrWithholdingAllowed,@JsonKey(includeIfNull: true, name: 'isr_withholding_rate') DecimalRate? isrWithholdingRate,@JsonKey(name: 'iva_withholding_allowed') bool ivaWithholdingAllowed,@JsonKey(includeIfNull: true, name: 'iva_withholding_rate') DecimalRate? ivaWithholdingRate,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity
});




}
/// @nodoc
class _$IncomeDefaultsCopyWithImpl<$Res>
    implements $IncomeDefaultsCopyWith<$Res> {
  _$IncomeDefaultsCopyWithImpl(this._self, this._then);

  final IncomeDefaults _self;
  final $Res Function(IncomeDefaults) _then;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ivaRatesAllowed = null,Object? suggestedIvaRate = null,Object? isrWithholdingAllowed = null,Object? isrWithholdingRate = freezed,Object? ivaWithholdingAllowed = null,Object? ivaWithholdingRate = freezed,Object? platformActivity = freezed,}) {
  return _then(IncomeDefaults(
ivaRatesAllowed: null == ivaRatesAllowed ? _self.ivaRatesAllowed : ivaRatesAllowed // ignore: cast_nullable_to_non_nullable
as List<IvaRate>,suggestedIvaRate: null == suggestedIvaRate ? _self.suggestedIvaRate : suggestedIvaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,isrWithholdingAllowed: null == isrWithholdingAllowed ? _self.isrWithholdingAllowed : isrWithholdingAllowed // ignore: cast_nullable_to_non_nullable
as bool,isrWithholdingRate: freezed == isrWithholdingRate ? _self.isrWithholdingRate : isrWithholdingRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,ivaWithholdingAllowed: null == ivaWithholdingAllowed ? _self.ivaWithholdingAllowed : ivaWithholdingAllowed // ignore: cast_nullable_to_non_nullable
as bool,ivaWithholdingRate: freezed == ivaWithholdingRate ? _self.ivaWithholdingRate : ivaWithholdingRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,
  ));
}

}


/// Adds pattern-matching-related methods to [IncomeDefaults].
extension IncomeDefaultsPatterns on IncomeDefaults {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomeDefaults value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomeDefaults value)  $default,){
final _that = this;
switch (_that) {
case _IncomeDefaults():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomeDefaults value)?  $default,){
final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'isr_withholding_allowed')  bool isrWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'isr_withholding_rate')  DecimalRate? isrWithholdingRate, @JsonKey(name: 'iva_withholding_allowed')  bool ivaWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'iva_withholding_rate')  DecimalRate? ivaWithholdingRate, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.isrWithholdingAllowed,_that.isrWithholdingRate,_that.ivaWithholdingAllowed,_that.ivaWithholdingRate,_that.platformActivity);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'isr_withholding_allowed')  bool isrWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'isr_withholding_rate')  DecimalRate? isrWithholdingRate, @JsonKey(name: 'iva_withholding_allowed')  bool ivaWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'iva_withholding_rate')  DecimalRate? ivaWithholdingRate, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)  $default,) {final _that = this;
switch (_that) {
case _IncomeDefaults():
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.isrWithholdingAllowed,_that.isrWithholdingRate,_that.ivaWithholdingAllowed,_that.ivaWithholdingRate,_that.platformActivity);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'isr_withholding_allowed')  bool isrWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'isr_withholding_rate')  DecimalRate? isrWithholdingRate, @JsonKey(name: 'iva_withholding_allowed')  bool ivaWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'iva_withholding_rate')  DecimalRate? ivaWithholdingRate, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)?  $default,) {final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.isrWithholdingAllowed,_that.isrWithholdingRate,_that.ivaWithholdingAllowed,_that.ivaWithholdingRate,_that.platformActivity);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeDefaults implements IncomeDefaults {
  const _IncomeDefaults({@JsonKey(name: 'iva_rates_allowed') required  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate') required this.suggestedIvaRate, @JsonKey(name: 'isr_withholding_allowed') required this.isrWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'isr_withholding_rate') required this.isrWithholdingRate, @JsonKey(name: 'iva_withholding_allowed') required this.ivaWithholdingAllowed, @JsonKey(includeIfNull: true, name: 'iva_withholding_rate') required this.ivaWithholdingRate, @JsonKey(includeIfNull: true, name: 'platform_activity') required this.platformActivity}): _ivaRatesAllowed = ivaRatesAllowed;
  factory _IncomeDefaults.fromJson(Map<String, dynamic> json) => _$IncomeDefaultsFromJson(json);

 final  List<IvaRate> _ivaRatesAllowed;
@override@JsonKey(name: 'iva_rates_allowed') List<IvaRate> get ivaRatesAllowed {
  if (_ivaRatesAllowed is EqualUnmodifiableListView) return _ivaRatesAllowed;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_ivaRatesAllowed);
}

@override@JsonKey(name: 'suggested_iva_rate') final  IvaRate suggestedIvaRate;
@override@JsonKey(name: 'isr_withholding_allowed') final  bool isrWithholdingAllowed;
/// Fraction of `amount_cents`; `null` when not allowed or when there is no suggestion.
@override@JsonKey(includeIfNull: true, name: 'isr_withholding_rate') final  DecimalRate? isrWithholdingRate;
@override@JsonKey(name: 'iva_withholding_allowed') final  bool ivaWithholdingAllowed;
/// Fraction of `iva_cents` (2/3 = `0.666667`); `null` when not allowed or when there is no suggestion.
@override@JsonKey(includeIfNull: true, name: 'iva_withholding_rate') final  DecimalRate? ivaWithholdingRate;
/// The 625 activity the ISR rate is for; `null` outside 625 or when it is ambiguous.
@override@JsonKey(includeIfNull: true, name: 'platform_activity') final  PlatformActivity? platformActivity;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeDefaultsCopyWith<_IncomeDefaults> get copyWith => __$IncomeDefaultsCopyWithImpl<_IncomeDefaults>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeDefaultsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeDefaults&&const DeepCollectionEquality().equals(other.ivaRatesAllowed, _ivaRatesAllowed)&&(identical(other.suggestedIvaRate, suggestedIvaRate) || other.suggestedIvaRate == suggestedIvaRate)&&(identical(other.isrWithholdingAllowed, isrWithholdingAllowed) || other.isrWithholdingAllowed == isrWithholdingAllowed)&&(identical(other.isrWithholdingRate, isrWithholdingRate) || other.isrWithholdingRate == isrWithholdingRate)&&(identical(other.ivaWithholdingAllowed, ivaWithholdingAllowed) || other.ivaWithholdingAllowed == ivaWithholdingAllowed)&&(identical(other.ivaWithholdingRate, ivaWithholdingRate) || other.ivaWithholdingRate == ivaWithholdingRate)&&(identical(other.platformActivity, platformActivity) || other.platformActivity == platformActivity));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_ivaRatesAllowed),suggestedIvaRate,isrWithholdingAllowed,isrWithholdingRate,ivaWithholdingAllowed,ivaWithholdingRate,platformActivity);
}

@override
String toString() {
    return 'IncomeDefaults(ivaRatesAllowed: $ivaRatesAllowed, suggestedIvaRate: $suggestedIvaRate, isrWithholdingAllowed: $isrWithholdingAllowed, isrWithholdingRate: $isrWithholdingRate, ivaWithholdingAllowed: $ivaWithholdingAllowed, ivaWithholdingRate: $ivaWithholdingRate, platformActivity: $platformActivity)';
}


}

/// @nodoc
abstract mixin class _$IncomeDefaultsCopyWith<$Res> implements $IncomeDefaultsCopyWith<$Res> {
  factory _$IncomeDefaultsCopyWith(_IncomeDefaults value, $Res Function(_IncomeDefaults) _then) = __$IncomeDefaultsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'iva_rates_allowed') List<IvaRate> ivaRatesAllowed,@JsonKey(name: 'suggested_iva_rate') IvaRate suggestedIvaRate,@JsonKey(name: 'isr_withholding_allowed') bool isrWithholdingAllowed,@JsonKey(includeIfNull: true, name: 'isr_withholding_rate') DecimalRate? isrWithholdingRate,@JsonKey(name: 'iva_withholding_allowed') bool ivaWithholdingAllowed,@JsonKey(includeIfNull: true, name: 'iva_withholding_rate') DecimalRate? ivaWithholdingRate,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity
});




}
/// @nodoc
class __$IncomeDefaultsCopyWithImpl<$Res>
    implements _$IncomeDefaultsCopyWith<$Res> {
  __$IncomeDefaultsCopyWithImpl(this._self, this._then);

  final _IncomeDefaults _self;
  final $Res Function(_IncomeDefaults) _then;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ivaRatesAllowed = null,Object? suggestedIvaRate = null,Object? isrWithholdingAllowed = null,Object? isrWithholdingRate = freezed,Object? ivaWithholdingAllowed = null,Object? ivaWithholdingRate = freezed,Object? platformActivity = freezed,}) {
  return _then(_IncomeDefaults(
ivaRatesAllowed: null == ivaRatesAllowed ? _self._ivaRatesAllowed : ivaRatesAllowed // ignore: cast_nullable_to_non_nullable
as List<IvaRate>,suggestedIvaRate: null == suggestedIvaRate ? _self.suggestedIvaRate : suggestedIvaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,isrWithholdingAllowed: null == isrWithholdingAllowed ? _self.isrWithholdingAllowed : isrWithholdingAllowed // ignore: cast_nullable_to_non_nullable
as bool,isrWithholdingRate: freezed == isrWithholdingRate ? _self.isrWithholdingRate : isrWithholdingRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,ivaWithholdingAllowed: null == ivaWithholdingAllowed ? _self.ivaWithholdingAllowed : ivaWithholdingAllowed // ignore: cast_nullable_to_non_nullable
as bool,ivaWithholdingRate: freezed == ivaWithholdingRate ? _self.ivaWithholdingRate : ivaWithholdingRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,
  ));
}


}

// dart format on
