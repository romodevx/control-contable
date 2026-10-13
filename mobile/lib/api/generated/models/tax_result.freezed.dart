// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_result.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxResult {

/// Exact result; negative = saldo a favor.
@JsonKey(name: 'amount_cents') MoneyCents get amountCents;/// Payment amount. Equals amount_cents (centavo precision) until the CPA confirms CFF art. 20 whole-peso rounding (ADR-0002 §4).
@JsonKey(name: 'payable_rounded_cents') MoneyCents get payableRoundedCents; TaxResultOutcome get outcome;
/// Create a copy of TaxResult
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxResultCopyWith<TaxResult> get copyWith => _$TaxResultCopyWithImpl<TaxResult>(this as TaxResult, _$identity);

  /// Serializes this TaxResult to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxResult;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxResult&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.payableRoundedCents, _this.payableRoundedCents) || other.payableRoundedCents == _this.payableRoundedCents)&&(identical(other.outcome, _this.outcome) || other.outcome == _this.outcome));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxResult;
  return Object.hash(runtimeType,_this.amountCents,_this.payableRoundedCents,_this.outcome);
}

@override
String toString() {
  final _this = this as TaxResult;
  return 'TaxResult(amountCents: ${_this.amountCents}, payableRoundedCents: ${_this.payableRoundedCents}, outcome: ${_this.outcome})';
}


}

/// @nodoc
abstract mixin class $TaxResultCopyWith<$Res>  {
  factory $TaxResultCopyWith(TaxResult value, $Res Function(TaxResult) _then) = _$TaxResultCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'amount_cents') MoneyCents amountCents,@JsonKey(name: 'payable_rounded_cents') MoneyCents payableRoundedCents, TaxResultOutcome outcome
});




}
/// @nodoc
class _$TaxResultCopyWithImpl<$Res>
    implements $TaxResultCopyWith<$Res> {
  _$TaxResultCopyWithImpl(this._self, this._then);

  final TaxResult _self;
  final $Res Function(TaxResult) _then;

/// Create a copy of TaxResult
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? amountCents = null,Object? payableRoundedCents = null,Object? outcome = null,}) {
  return _then(TaxResult(
amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,payableRoundedCents: null == payableRoundedCents ? _self.payableRoundedCents : payableRoundedCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,outcome: null == outcome ? _self.outcome : outcome // ignore: cast_nullable_to_non_nullable
as TaxResultOutcome,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxResult].
extension TaxResultPatterns on TaxResult {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxResult value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxResult() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxResult value)  $default,){
final _that = this;
switch (_that) {
case _TaxResult():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxResult value)?  $default,){
final _that = this;
switch (_that) {
case _TaxResult() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(name: 'payable_rounded_cents')  MoneyCents payableRoundedCents,  TaxResultOutcome outcome)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxResult() when $default != null:
return $default(_that.amountCents,_that.payableRoundedCents,_that.outcome);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(name: 'payable_rounded_cents')  MoneyCents payableRoundedCents,  TaxResultOutcome outcome)  $default,) {final _that = this;
switch (_that) {
case _TaxResult():
return $default(_that.amountCents,_that.payableRoundedCents,_that.outcome);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'amount_cents')  MoneyCents amountCents, @JsonKey(name: 'payable_rounded_cents')  MoneyCents payableRoundedCents,  TaxResultOutcome outcome)?  $default,) {final _that = this;
switch (_that) {
case _TaxResult() when $default != null:
return $default(_that.amountCents,_that.payableRoundedCents,_that.outcome);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxResult implements TaxResult {
  const _TaxResult({@JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'payable_rounded_cents') required this.payableRoundedCents, required this.outcome});
  factory _TaxResult.fromJson(Map<String, dynamic> json) => _$TaxResultFromJson(json);

/// Exact result; negative = saldo a favor.
@override@JsonKey(name: 'amount_cents') final  MoneyCents amountCents;
/// Payment amount. Equals amount_cents (centavo precision) until the CPA confirms CFF art. 20 whole-peso rounding (ADR-0002 §4).
@override@JsonKey(name: 'payable_rounded_cents') final  MoneyCents payableRoundedCents;
@override final  TaxResultOutcome outcome;

/// Create a copy of TaxResult
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxResultCopyWith<_TaxResult> get copyWith => __$TaxResultCopyWithImpl<_TaxResult>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxResultToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxResult&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.payableRoundedCents, payableRoundedCents) || other.payableRoundedCents == payableRoundedCents)&&(identical(other.outcome, outcome) || other.outcome == outcome));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,amountCents,payableRoundedCents,outcome);
}

@override
String toString() {
    return 'TaxResult(amountCents: $amountCents, payableRoundedCents: $payableRoundedCents, outcome: $outcome)';
}


}

/// @nodoc
abstract mixin class _$TaxResultCopyWith<$Res> implements $TaxResultCopyWith<$Res> {
  factory _$TaxResultCopyWith(_TaxResult value, $Res Function(_TaxResult) _then) = __$TaxResultCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'amount_cents') MoneyCents amountCents,@JsonKey(name: 'payable_rounded_cents') MoneyCents payableRoundedCents, TaxResultOutcome outcome
});




}
/// @nodoc
class __$TaxResultCopyWithImpl<$Res>
    implements _$TaxResultCopyWith<$Res> {
  __$TaxResultCopyWithImpl(this._self, this._then);

  final _TaxResult _self;
  final $Res Function(_TaxResult) _then;

/// Create a copy of TaxResult
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? amountCents = null,Object? payableRoundedCents = null,Object? outcome = null,}) {
  return _then(_TaxResult(
amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,payableRoundedCents: null == payableRoundedCents ? _self.payableRoundedCents : payableRoundedCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,outcome: null == outcome ? _self.outcome : outcome // ignore: cast_nullable_to_non_nullable
as TaxResultOutcome,
  ));
}


}

// dart format on
