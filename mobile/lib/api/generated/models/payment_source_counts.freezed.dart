// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'payment_source_counts.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PaymentSourceCounts {

@JsonKey(name: 'filed_months') int get filedMonths;@JsonKey(name: 'tax_input_months') int get taxInputMonths;@JsonKey(name: 'estimated_by_app_months') int get estimatedByAppMonths;
/// Create a copy of PaymentSourceCounts
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PaymentSourceCountsCopyWith<PaymentSourceCounts> get copyWith => _$PaymentSourceCountsCopyWithImpl<PaymentSourceCounts>(this as PaymentSourceCounts, _$identity);

  /// Serializes this PaymentSourceCounts to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PaymentSourceCounts;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PaymentSourceCounts&&(identical(other.filedMonths, _this.filedMonths) || other.filedMonths == _this.filedMonths)&&(identical(other.taxInputMonths, _this.taxInputMonths) || other.taxInputMonths == _this.taxInputMonths)&&(identical(other.estimatedByAppMonths, _this.estimatedByAppMonths) || other.estimatedByAppMonths == _this.estimatedByAppMonths));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PaymentSourceCounts;
  return Object.hash(runtimeType,_this.filedMonths,_this.taxInputMonths,_this.estimatedByAppMonths);
}

@override
String toString() {
  final _this = this as PaymentSourceCounts;
  return 'PaymentSourceCounts(filedMonths: ${_this.filedMonths}, taxInputMonths: ${_this.taxInputMonths}, estimatedByAppMonths: ${_this.estimatedByAppMonths})';
}


}

/// @nodoc
abstract mixin class $PaymentSourceCountsCopyWith<$Res>  {
  factory $PaymentSourceCountsCopyWith(PaymentSourceCounts value, $Res Function(PaymentSourceCounts) _then) = _$PaymentSourceCountsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'filed_months') int filedMonths,@JsonKey(name: 'tax_input_months') int taxInputMonths,@JsonKey(name: 'estimated_by_app_months') int estimatedByAppMonths
});




}
/// @nodoc
class _$PaymentSourceCountsCopyWithImpl<$Res>
    implements $PaymentSourceCountsCopyWith<$Res> {
  _$PaymentSourceCountsCopyWithImpl(this._self, this._then);

  final PaymentSourceCounts _self;
  final $Res Function(PaymentSourceCounts) _then;

/// Create a copy of PaymentSourceCounts
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? filedMonths = null,Object? taxInputMonths = null,Object? estimatedByAppMonths = null,}) {
  return _then(PaymentSourceCounts(
filedMonths: null == filedMonths ? _self.filedMonths : filedMonths // ignore: cast_nullable_to_non_nullable
as int,taxInputMonths: null == taxInputMonths ? _self.taxInputMonths : taxInputMonths // ignore: cast_nullable_to_non_nullable
as int,estimatedByAppMonths: null == estimatedByAppMonths ? _self.estimatedByAppMonths : estimatedByAppMonths // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [PaymentSourceCounts].
extension PaymentSourceCountsPatterns on PaymentSourceCounts {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PaymentSourceCounts value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PaymentSourceCounts() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PaymentSourceCounts value)  $default,){
final _that = this;
switch (_that) {
case _PaymentSourceCounts():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PaymentSourceCounts value)?  $default,){
final _that = this;
switch (_that) {
case _PaymentSourceCounts() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'filed_months')  int filedMonths, @JsonKey(name: 'tax_input_months')  int taxInputMonths, @JsonKey(name: 'estimated_by_app_months')  int estimatedByAppMonths)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PaymentSourceCounts() when $default != null:
return $default(_that.filedMonths,_that.taxInputMonths,_that.estimatedByAppMonths);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'filed_months')  int filedMonths, @JsonKey(name: 'tax_input_months')  int taxInputMonths, @JsonKey(name: 'estimated_by_app_months')  int estimatedByAppMonths)  $default,) {final _that = this;
switch (_that) {
case _PaymentSourceCounts():
return $default(_that.filedMonths,_that.taxInputMonths,_that.estimatedByAppMonths);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'filed_months')  int filedMonths, @JsonKey(name: 'tax_input_months')  int taxInputMonths, @JsonKey(name: 'estimated_by_app_months')  int estimatedByAppMonths)?  $default,) {final _that = this;
switch (_that) {
case _PaymentSourceCounts() when $default != null:
return $default(_that.filedMonths,_that.taxInputMonths,_that.estimatedByAppMonths);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PaymentSourceCounts implements PaymentSourceCounts {
  const _PaymentSourceCounts({@JsonKey(name: 'filed_months') required this.filedMonths, @JsonKey(name: 'tax_input_months') required this.taxInputMonths, @JsonKey(name: 'estimated_by_app_months') required this.estimatedByAppMonths});
  factory _PaymentSourceCounts.fromJson(Map<String, dynamic> json) => _$PaymentSourceCountsFromJson(json);

@override@JsonKey(name: 'filed_months') final  int filedMonths;
@override@JsonKey(name: 'tax_input_months') final  int taxInputMonths;
@override@JsonKey(name: 'estimated_by_app_months') final  int estimatedByAppMonths;

/// Create a copy of PaymentSourceCounts
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PaymentSourceCountsCopyWith<_PaymentSourceCounts> get copyWith => __$PaymentSourceCountsCopyWithImpl<_PaymentSourceCounts>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PaymentSourceCountsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PaymentSourceCounts&&(identical(other.filedMonths, filedMonths) || other.filedMonths == filedMonths)&&(identical(other.taxInputMonths, taxInputMonths) || other.taxInputMonths == taxInputMonths)&&(identical(other.estimatedByAppMonths, estimatedByAppMonths) || other.estimatedByAppMonths == estimatedByAppMonths));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,filedMonths,taxInputMonths,estimatedByAppMonths);
}

@override
String toString() {
    return 'PaymentSourceCounts(filedMonths: $filedMonths, taxInputMonths: $taxInputMonths, estimatedByAppMonths: $estimatedByAppMonths)';
}


}

/// @nodoc
abstract mixin class _$PaymentSourceCountsCopyWith<$Res> implements $PaymentSourceCountsCopyWith<$Res> {
  factory _$PaymentSourceCountsCopyWith(_PaymentSourceCounts value, $Res Function(_PaymentSourceCounts) _then) = __$PaymentSourceCountsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'filed_months') int filedMonths,@JsonKey(name: 'tax_input_months') int taxInputMonths,@JsonKey(name: 'estimated_by_app_months') int estimatedByAppMonths
});




}
/// @nodoc
class __$PaymentSourceCountsCopyWithImpl<$Res>
    implements _$PaymentSourceCountsCopyWith<$Res> {
  __$PaymentSourceCountsCopyWithImpl(this._self, this._then);

  final _PaymentSourceCounts _self;
  final $Res Function(_PaymentSourceCounts) _then;

/// Create a copy of PaymentSourceCounts
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? filedMonths = null,Object? taxInputMonths = null,Object? estimatedByAppMonths = null,}) {
  return _then(_PaymentSourceCounts(
filedMonths: null == filedMonths ? _self.filedMonths : filedMonths // ignore: cast_nullable_to_non_nullable
as int,taxInputMonths: null == taxInputMonths ? _self.taxInputMonths : taxInputMonths // ignore: cast_nullable_to_non_nullable
as int,estimatedByAppMonths: null == estimatedByAppMonths ? _self.estimatedByAppMonths : estimatedByAppMonths // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
