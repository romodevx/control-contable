// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'trend_point.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TrendPoint {

 Period get period;@JsonKey(name: 'income_cents') NonNegativeCents get incomeCents;@JsonKey(name: 'expense_cents') NonNegativeCents get expenseCents;@JsonKey(name: 'tax_payable_rounded_cents') NonNegativeCents get taxPayableRoundedCents;
/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TrendPointCopyWith<TrendPoint> get copyWith => _$TrendPointCopyWithImpl<TrendPoint>(this as TrendPoint, _$identity);

  /// Serializes this TrendPoint to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TrendPoint;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TrendPoint&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.incomeCents, _this.incomeCents) || other.incomeCents == _this.incomeCents)&&(identical(other.expenseCents, _this.expenseCents) || other.expenseCents == _this.expenseCents)&&(identical(other.taxPayableRoundedCents, _this.taxPayableRoundedCents) || other.taxPayableRoundedCents == _this.taxPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TrendPoint;
  return Object.hash(runtimeType,_this.period,_this.incomeCents,_this.expenseCents,_this.taxPayableRoundedCents);
}

@override
String toString() {
  final _this = this as TrendPoint;
  return 'TrendPoint(period: ${_this.period}, incomeCents: ${_this.incomeCents}, expenseCents: ${_this.expenseCents}, taxPayableRoundedCents: ${_this.taxPayableRoundedCents})';
}


}

/// @nodoc
abstract mixin class $TrendPointCopyWith<$Res>  {
  factory $TrendPointCopyWith(TrendPoint value, $Res Function(TrendPoint) _then) = _$TrendPointCopyWithImpl;
@useResult
$Res call({
 Period period,@JsonKey(name: 'income_cents') NonNegativeCents incomeCents,@JsonKey(name: 'expense_cents') NonNegativeCents expenseCents,@JsonKey(name: 'tax_payable_rounded_cents') NonNegativeCents taxPayableRoundedCents
});




}
/// @nodoc
class _$TrendPointCopyWithImpl<$Res>
    implements $TrendPointCopyWith<$Res> {
  _$TrendPointCopyWithImpl(this._self, this._then);

  final TrendPoint _self;
  final $Res Function(TrendPoint) _then;

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? period = null,Object? incomeCents = null,Object? expenseCents = null,Object? taxPayableRoundedCents = null,}) {
  return _then(TrendPoint(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,incomeCents: null == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,expenseCents: null == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,taxPayableRoundedCents: null == taxPayableRoundedCents ? _self.taxPayableRoundedCents : taxPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [TrendPoint].
extension TrendPointPatterns on TrendPoint {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TrendPoint value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TrendPoint value)  $default,){
final _that = this;
switch (_that) {
case _TrendPoint():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TrendPoint value)?  $default,){
final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Period period, @JsonKey(name: 'income_cents')  NonNegativeCents incomeCents, @JsonKey(name: 'expense_cents')  NonNegativeCents expenseCents, @JsonKey(name: 'tax_payable_rounded_cents')  NonNegativeCents taxPayableRoundedCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.taxPayableRoundedCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Period period, @JsonKey(name: 'income_cents')  NonNegativeCents incomeCents, @JsonKey(name: 'expense_cents')  NonNegativeCents expenseCents, @JsonKey(name: 'tax_payable_rounded_cents')  NonNegativeCents taxPayableRoundedCents)  $default,) {final _that = this;
switch (_that) {
case _TrendPoint():
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.taxPayableRoundedCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Period period, @JsonKey(name: 'income_cents')  NonNegativeCents incomeCents, @JsonKey(name: 'expense_cents')  NonNegativeCents expenseCents, @JsonKey(name: 'tax_payable_rounded_cents')  NonNegativeCents taxPayableRoundedCents)?  $default,) {final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.taxPayableRoundedCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TrendPoint implements TrendPoint {
  const _TrendPoint({required this.period, @JsonKey(name: 'income_cents') required this.incomeCents, @JsonKey(name: 'expense_cents') required this.expenseCents, @JsonKey(name: 'tax_payable_rounded_cents') required this.taxPayableRoundedCents});
  factory _TrendPoint.fromJson(Map<String, dynamic> json) => _$TrendPointFromJson(json);

@override final  Period period;
@override@JsonKey(name: 'income_cents') final  NonNegativeCents incomeCents;
@override@JsonKey(name: 'expense_cents') final  NonNegativeCents expenseCents;
@override@JsonKey(name: 'tax_payable_rounded_cents') final  NonNegativeCents taxPayableRoundedCents;

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TrendPointCopyWith<_TrendPoint> get copyWith => __$TrendPointCopyWithImpl<_TrendPoint>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TrendPointToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TrendPoint&&(identical(other.period, period) || other.period == period)&&(identical(other.incomeCents, incomeCents) || other.incomeCents == incomeCents)&&(identical(other.expenseCents, expenseCents) || other.expenseCents == expenseCents)&&(identical(other.taxPayableRoundedCents, taxPayableRoundedCents) || other.taxPayableRoundedCents == taxPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,period,incomeCents,expenseCents,taxPayableRoundedCents);
}

@override
String toString() {
    return 'TrendPoint(period: $period, incomeCents: $incomeCents, expenseCents: $expenseCents, taxPayableRoundedCents: $taxPayableRoundedCents)';
}


}

/// @nodoc
abstract mixin class _$TrendPointCopyWith<$Res> implements $TrendPointCopyWith<$Res> {
  factory _$TrendPointCopyWith(_TrendPoint value, $Res Function(_TrendPoint) _then) = __$TrendPointCopyWithImpl;
@override @useResult
$Res call({
 Period period,@JsonKey(name: 'income_cents') NonNegativeCents incomeCents,@JsonKey(name: 'expense_cents') NonNegativeCents expenseCents,@JsonKey(name: 'tax_payable_rounded_cents') NonNegativeCents taxPayableRoundedCents
});




}
/// @nodoc
class __$TrendPointCopyWithImpl<$Res>
    implements _$TrendPointCopyWith<$Res> {
  __$TrendPointCopyWithImpl(this._self, this._then);

  final _TrendPoint _self;
  final $Res Function(_TrendPoint) _then;

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? period = null,Object? incomeCents = null,Object? expenseCents = null,Object? taxPayableRoundedCents = null,}) {
  return _then(_TrendPoint(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,incomeCents: null == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,expenseCents: null == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,taxPayableRoundedCents: null == taxPayableRoundedCents ? _self.taxPayableRoundedCents : taxPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
