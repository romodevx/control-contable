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

 Period get period;@JsonKey(includeIfNull: true, name: 'income_cents') int? get incomeCents;@JsonKey(includeIfNull: true, name: 'expense_cents') int? get expenseCents;@JsonKey(includeIfNull: true) TrendTax? get tax;
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
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TrendPoint&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.incomeCents, _this.incomeCents) || other.incomeCents == _this.incomeCents)&&(identical(other.expenseCents, _this.expenseCents) || other.expenseCents == _this.expenseCents)&&(identical(other.tax, _this.tax) || other.tax == _this.tax));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TrendPoint;
  return Object.hash(runtimeType,_this.period,_this.incomeCents,_this.expenseCents,_this.tax);
}

@override
String toString() {
  final _this = this as TrendPoint;
  return 'TrendPoint(period: ${_this.period}, incomeCents: ${_this.incomeCents}, expenseCents: ${_this.expenseCents}, tax: ${_this.tax})';
}


}

/// @nodoc
abstract mixin class $TrendPointCopyWith<$Res>  {
  factory $TrendPointCopyWith(TrendPoint value, $Res Function(TrendPoint) _then) = _$TrendPointCopyWithImpl;
@useResult
$Res call({
 Period period,@JsonKey(includeIfNull: true, name: 'income_cents') int? incomeCents,@JsonKey(includeIfNull: true, name: 'expense_cents') int? expenseCents,@JsonKey(includeIfNull: true) TrendTax? tax
});


$TrendTaxCopyWith<$Res>? get tax;

}
/// @nodoc
class _$TrendPointCopyWithImpl<$Res>
    implements $TrendPointCopyWith<$Res> {
  _$TrendPointCopyWithImpl(this._self, this._then);

  final TrendPoint _self;
  final $Res Function(TrendPoint) _then;

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? period = null,Object? incomeCents = freezed,Object? expenseCents = freezed,Object? tax = freezed,}) {
  return _then(TrendPoint(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,incomeCents: freezed == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as int?,expenseCents: freezed == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as int?,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as TrendTax?,
  ));
}
/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TrendTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $TrendTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Period period, @JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents, @JsonKey(includeIfNull: true)  TrendTax? tax)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.tax);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Period period, @JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents, @JsonKey(includeIfNull: true)  TrendTax? tax)  $default,) {final _that = this;
switch (_that) {
case _TrendPoint():
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.tax);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Period period, @JsonKey(includeIfNull: true, name: 'income_cents')  int? incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents')  int? expenseCents, @JsonKey(includeIfNull: true)  TrendTax? tax)?  $default,) {final _that = this;
switch (_that) {
case _TrendPoint() when $default != null:
return $default(_that.period,_that.incomeCents,_that.expenseCents,_that.tax);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TrendPoint implements TrendPoint {
  const _TrendPoint({required this.period, @JsonKey(includeIfNull: true, name: 'income_cents') required this.incomeCents, @JsonKey(includeIfNull: true, name: 'expense_cents') required this.expenseCents, @JsonKey(includeIfNull: true) required this.tax});
  factory _TrendPoint.fromJson(Map<String, dynamic> json) => _$TrendPointFromJson(json);

@override final  Period period;
@override@JsonKey(includeIfNull: true, name: 'income_cents') final  int? incomeCents;
@override@JsonKey(includeIfNull: true, name: 'expense_cents') final  int? expenseCents;
@override@JsonKey(includeIfNull: true) final  TrendTax? tax;

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
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TrendPoint&&(identical(other.period, period) || other.period == period)&&(identical(other.incomeCents, incomeCents) || other.incomeCents == incomeCents)&&(identical(other.expenseCents, expenseCents) || other.expenseCents == expenseCents)&&(identical(other.tax, tax) || other.tax == tax));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,period,incomeCents,expenseCents,tax);
}

@override
String toString() {
    return 'TrendPoint(period: $period, incomeCents: $incomeCents, expenseCents: $expenseCents, tax: $tax)';
}


}

/// @nodoc
abstract mixin class _$TrendPointCopyWith<$Res> implements $TrendPointCopyWith<$Res> {
  factory _$TrendPointCopyWith(_TrendPoint value, $Res Function(_TrendPoint) _then) = __$TrendPointCopyWithImpl;
@override @useResult
$Res call({
 Period period,@JsonKey(includeIfNull: true, name: 'income_cents') int? incomeCents,@JsonKey(includeIfNull: true, name: 'expense_cents') int? expenseCents,@JsonKey(includeIfNull: true) TrendTax? tax
});


@override $TrendTaxCopyWith<$Res>? get tax;

}
/// @nodoc
class __$TrendPointCopyWithImpl<$Res>
    implements _$TrendPointCopyWith<$Res> {
  __$TrendPointCopyWithImpl(this._self, this._then);

  final _TrendPoint _self;
  final $Res Function(_TrendPoint) _then;

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? period = null,Object? incomeCents = freezed,Object? expenseCents = freezed,Object? tax = freezed,}) {
  return _then(_TrendPoint(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,incomeCents: freezed == incomeCents ? _self.incomeCents : incomeCents // ignore: cast_nullable_to_non_nullable
as int?,expenseCents: freezed == expenseCents ? _self.expenseCents : expenseCents // ignore: cast_nullable_to_non_nullable
as int?,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as TrendTax?,
  ));
}

/// Create a copy of TrendPoint
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TrendTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $TrendTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
}
}

// dart format on
