// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'resico_pf_check.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ResicoPfCheck {

@JsonKey(name: 'income_cents') NonNegativeCents get incomeCents; DecimalRate get rate;@JsonKey(name: 'annual_isr_cents') NonNegativeCents get annualIsrCents;@JsonKey(name: 'monthly_payments_cents') NonNegativeCents get monthlyPaymentsCents;@JsonKey(name: 'withheld_cents') NonNegativeCents get withheldCents;@JsonKey(name: 'difference_cents') MoneyCents get differenceCents;
/// Create a copy of ResicoPfCheck
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ResicoPfCheckCopyWith<ResicoPfCheck> get copyWith => _$ResicoPfCheckCopyWithImpl<ResicoPfCheck>(this as ResicoPfCheck, _$identity);

  /// Serializes this ResicoPfCheck to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ResicoPfCheck;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ResicoPfCheck&&(identical(other.incomeCents, _this.incomeCents) || other.incomeCents == _this.incomeCents)&&(identical(other.rate, _this.rate) || other.rate == _this.rate)&&(identical(other.annualIsrCents, _this.annualIsrCents) || other.annualIsrCents == _this.annualIsrCents)&&(identical(other.monthlyPaymentsCents, _this.monthlyPaymentsCents) || other.monthlyPaymentsCents == _this.monthlyPaymentsCents)&&(identical(other.withheldCents, _this.withheldCents) || other.withheldCents == _this.withheldCents)&&(identical(other.differenceCents, _this.differenceCents) || other.differenceCents == _this.differenceCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ResicoPfCheck;
  return Object.hash(runtimeType,_this.incomeCents,_this.rate,_this.annualIsrCents,_this.monthlyPaymentsCents,_this.withheldCents,_this.differenceCents);
}

@override
String toString() {
  final _this = this as ResicoPfCheck;
  return 'ResicoPfCheck(incomeCents: ${_this.incomeCents}, rate: ${_this.rate}, annualIsrCents: ${_this.annualIsrCents}, monthlyPaymentsCents: ${_this.monthlyPaymentsCents}, withheldCents: ${_this.withheldCents}, differenceCents: ${_this.differenceCents})';
}


}

/// @nodoc
abstract mixin class $ResicoPfCheckCopyWith<$Res>  {
  factory $ResicoPfCheckCopyWith(ResicoPfCheck value, $Res Function(ResicoPfCheck) _then) = _$ResicoPfCheckCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'income_cents') NonNegativeCents incomeCents, DecimalRate rate,@JsonKey(name: 'annual_isr_cents') NonNegativeCents annualIsrCents,@JsonKey(name: 'monthly_payments_cents') NonNegativeCents monthlyPaymentsCents,@JsonKey(name: 'withheld_cents') NonNegativeCents withheldCents,@JsonKey(name: 'difference_cents') MoneyCents differenceCents
});




}
/// @nodoc
class _$ResicoPfCheckCopyWithImpl<$Res>
    implements $ResicoPfCheckCopyWith<$Res> {
  _$ResicoPfCheckCopyWithImpl(this._self, this._then);

  final ResicoPfCheck _self;
  final $Res Function(ResicoPfCheck) _then;

/// Create a copy of ResicoPfCheck
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? incomeCents = null,Object? rate = null,Object? annualIsrCents = null,Object? monthlyPaymentsCents = null,Object? withheldCents = null,Object? differenceCents = null,}) {
  return _then(ResicoPfCheck(
incomeCents: null == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,rate: null == rate ? _self.rate : rate // ignore: cast_nullable_to_non_nullable
as DecimalRate,annualIsrCents: null == annualIsrCents ? _self.annualIsrCents : annualIsrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,monthlyPaymentsCents: null == monthlyPaymentsCents ? _self.monthlyPaymentsCents : monthlyPaymentsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withheldCents: null == withheldCents ? _self.withheldCents : withheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,differenceCents: null == differenceCents ? _self.differenceCents : differenceCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,
  ));
}

}


/// Adds pattern-matching-related methods to [ResicoPfCheck].
extension ResicoPfCheckPatterns on ResicoPfCheck {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ResicoPfCheck value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ResicoPfCheck() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ResicoPfCheck value)  $default,){
final _that = this;
switch (_that) {
case _ResicoPfCheck():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ResicoPfCheck value)?  $default,){
final _that = this;
switch (_that) {
case _ResicoPfCheck() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'income_cents')  NonNegativeCents incomeCents,  DecimalRate rate, @JsonKey(name: 'annual_isr_cents')  NonNegativeCents annualIsrCents, @JsonKey(name: 'monthly_payments_cents')  NonNegativeCents monthlyPaymentsCents, @JsonKey(name: 'withheld_cents')  NonNegativeCents withheldCents, @JsonKey(name: 'difference_cents')  MoneyCents differenceCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ResicoPfCheck() when $default != null:
return $default(_that.incomeCents,_that.rate,_that.annualIsrCents,_that.monthlyPaymentsCents,_that.withheldCents,_that.differenceCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'income_cents')  NonNegativeCents incomeCents,  DecimalRate rate, @JsonKey(name: 'annual_isr_cents')  NonNegativeCents annualIsrCents, @JsonKey(name: 'monthly_payments_cents')  NonNegativeCents monthlyPaymentsCents, @JsonKey(name: 'withheld_cents')  NonNegativeCents withheldCents, @JsonKey(name: 'difference_cents')  MoneyCents differenceCents)  $default,) {final _that = this;
switch (_that) {
case _ResicoPfCheck():
return $default(_that.incomeCents,_that.rate,_that.annualIsrCents,_that.monthlyPaymentsCents,_that.withheldCents,_that.differenceCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'income_cents')  NonNegativeCents incomeCents,  DecimalRate rate, @JsonKey(name: 'annual_isr_cents')  NonNegativeCents annualIsrCents, @JsonKey(name: 'monthly_payments_cents')  NonNegativeCents monthlyPaymentsCents, @JsonKey(name: 'withheld_cents')  NonNegativeCents withheldCents, @JsonKey(name: 'difference_cents')  MoneyCents differenceCents)?  $default,) {final _that = this;
switch (_that) {
case _ResicoPfCheck() when $default != null:
return $default(_that.incomeCents,_that.rate,_that.annualIsrCents,_that.monthlyPaymentsCents,_that.withheldCents,_that.differenceCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ResicoPfCheck implements ResicoPfCheck {
  const _ResicoPfCheck({@JsonKey(name: 'income_cents') required this.incomeCents, required this.rate, @JsonKey(name: 'annual_isr_cents') required this.annualIsrCents, @JsonKey(name: 'monthly_payments_cents') required this.monthlyPaymentsCents, @JsonKey(name: 'withheld_cents') required this.withheldCents, @JsonKey(name: 'difference_cents') required this.differenceCents});
  factory _ResicoPfCheck.fromJson(Map<String, dynamic> json) => _$ResicoPfCheckFromJson(json);

@override@JsonKey(name: 'income_cents') final  NonNegativeCents incomeCents;
@override final  DecimalRate rate;
@override@JsonKey(name: 'annual_isr_cents') final  NonNegativeCents annualIsrCents;
@override@JsonKey(name: 'monthly_payments_cents') final  NonNegativeCents monthlyPaymentsCents;
@override@JsonKey(name: 'withheld_cents') final  NonNegativeCents withheldCents;
@override@JsonKey(name: 'difference_cents') final  MoneyCents differenceCents;

/// Create a copy of ResicoPfCheck
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ResicoPfCheckCopyWith<_ResicoPfCheck> get copyWith => __$ResicoPfCheckCopyWithImpl<_ResicoPfCheck>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ResicoPfCheckToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ResicoPfCheck&&(identical(other.incomeCents, incomeCents) || other.incomeCents == incomeCents)&&(identical(other.rate, rate) || other.rate == rate)&&(identical(other.annualIsrCents, annualIsrCents) || other.annualIsrCents == annualIsrCents)&&(identical(other.monthlyPaymentsCents, monthlyPaymentsCents) || other.monthlyPaymentsCents == monthlyPaymentsCents)&&(identical(other.withheldCents, withheldCents) || other.withheldCents == withheldCents)&&(identical(other.differenceCents, differenceCents) || other.differenceCents == differenceCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,incomeCents,rate,annualIsrCents,monthlyPaymentsCents,withheldCents,differenceCents);
}

@override
String toString() {
    return 'ResicoPfCheck(incomeCents: $incomeCents, rate: $rate, annualIsrCents: $annualIsrCents, monthlyPaymentsCents: $monthlyPaymentsCents, withheldCents: $withheldCents, differenceCents: $differenceCents)';
}


}

/// @nodoc
abstract mixin class _$ResicoPfCheckCopyWith<$Res> implements $ResicoPfCheckCopyWith<$Res> {
  factory _$ResicoPfCheckCopyWith(_ResicoPfCheck value, $Res Function(_ResicoPfCheck) _then) = __$ResicoPfCheckCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'income_cents') NonNegativeCents incomeCents, DecimalRate rate,@JsonKey(name: 'annual_isr_cents') NonNegativeCents annualIsrCents,@JsonKey(name: 'monthly_payments_cents') NonNegativeCents monthlyPaymentsCents,@JsonKey(name: 'withheld_cents') NonNegativeCents withheldCents,@JsonKey(name: 'difference_cents') MoneyCents differenceCents
});




}
/// @nodoc
class __$ResicoPfCheckCopyWithImpl<$Res>
    implements _$ResicoPfCheckCopyWith<$Res> {
  __$ResicoPfCheckCopyWithImpl(this._self, this._then);

  final _ResicoPfCheck _self;
  final $Res Function(_ResicoPfCheck) _then;

/// Create a copy of ResicoPfCheck
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? incomeCents = null,Object? rate = null,Object? annualIsrCents = null,Object? monthlyPaymentsCents = null,Object? withheldCents = null,Object? differenceCents = null,}) {
  return _then(_ResicoPfCheck(
incomeCents: null == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,rate: null == rate ? _self.rate : rate // ignore: cast_nullable_to_non_nullable
as DecimalRate,annualIsrCents: null == annualIsrCents ? _self.annualIsrCents : annualIsrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,monthlyPaymentsCents: null == monthlyPaymentsCents ? _self.monthlyPaymentsCents : monthlyPaymentsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withheldCents: null == withheldCents ? _self.withheldCents : withheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,differenceCents: null == differenceCents ? _self.differenceCents : differenceCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,
  ));
}


}

// dart format on
