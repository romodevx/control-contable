// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_input_value.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxInputValue {

 String get key;/// Month of the fiscal year, only for `money_monthly`.
@JsonKey(includeIfNull: false) int? get month;@JsonKey(includeIfNull: false, name: 'value_cents') NonNegativeCents? get valueCents;@JsonKey(includeIfNull: false, name: 'value_rate') DecimalRate? get valueRate;@JsonKey(includeIfNull: false, name: 'value_bool') bool? get valueBool;
/// Create a copy of TaxInputValue
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxInputValueCopyWith<TaxInputValue> get copyWith => _$TaxInputValueCopyWithImpl<TaxInputValue>(this as TaxInputValue, _$identity);

  /// Serializes this TaxInputValue to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxInputValue;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxInputValue&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.month, _this.month) || other.month == _this.month)&&(identical(other.valueCents, _this.valueCents) || other.valueCents == _this.valueCents)&&(identical(other.valueRate, _this.valueRate) || other.valueRate == _this.valueRate)&&(identical(other.valueBool, _this.valueBool) || other.valueBool == _this.valueBool));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxInputValue;
  return Object.hash(runtimeType,_this.key,_this.month,_this.valueCents,_this.valueRate,_this.valueBool);
}

@override
String toString() {
  final _this = this as TaxInputValue;
  return 'TaxInputValue(key: ${_this.key}, month: ${_this.month}, valueCents: ${_this.valueCents}, valueRate: ${_this.valueRate}, valueBool: ${_this.valueBool})';
}


}

/// @nodoc
abstract mixin class $TaxInputValueCopyWith<$Res>  {
  factory $TaxInputValueCopyWith(TaxInputValue value, $Res Function(TaxInputValue) _then) = _$TaxInputValueCopyWithImpl;
@useResult
$Res call({
 String key,@JsonKey(includeIfNull: false) int? month,@JsonKey(includeIfNull: false, name: 'value_cents') NonNegativeCents? valueCents,@JsonKey(includeIfNull: false, name: 'value_rate') DecimalRate? valueRate,@JsonKey(includeIfNull: false, name: 'value_bool') bool? valueBool
});




}
/// @nodoc
class _$TaxInputValueCopyWithImpl<$Res>
    implements $TaxInputValueCopyWith<$Res> {
  _$TaxInputValueCopyWithImpl(this._self, this._then);

  final TaxInputValue _self;
  final $Res Function(TaxInputValue) _then;

/// Create a copy of TaxInputValue
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? key = null,Object? month = freezed,Object? valueCents = freezed,Object? valueRate = freezed,Object? valueBool = freezed,}) {
  return _then(TaxInputValue(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,month: freezed == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int?,valueCents: freezed == valueCents ? _self.valueCents : valueCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,valueRate: freezed == valueRate ? _self.valueRate : valueRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,valueBool: freezed == valueBool ? _self.valueBool : valueBool // ignore: cast_nullable_to_non_nullable
as bool?,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxInputValue].
extension TaxInputValuePatterns on TaxInputValue {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxInputValue value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxInputValue() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxInputValue value)  $default,){
final _that = this;
switch (_that) {
case _TaxInputValue():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxInputValue value)?  $default,){
final _that = this;
switch (_that) {
case _TaxInputValue() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String key, @JsonKey(includeIfNull: false)  int? month, @JsonKey(includeIfNull: false, name: 'value_cents')  NonNegativeCents? valueCents, @JsonKey(includeIfNull: false, name: 'value_rate')  DecimalRate? valueRate, @JsonKey(includeIfNull: false, name: 'value_bool')  bool? valueBool)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxInputValue() when $default != null:
return $default(_that.key,_that.month,_that.valueCents,_that.valueRate,_that.valueBool);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String key, @JsonKey(includeIfNull: false)  int? month, @JsonKey(includeIfNull: false, name: 'value_cents')  NonNegativeCents? valueCents, @JsonKey(includeIfNull: false, name: 'value_rate')  DecimalRate? valueRate, @JsonKey(includeIfNull: false, name: 'value_bool')  bool? valueBool)  $default,) {final _that = this;
switch (_that) {
case _TaxInputValue():
return $default(_that.key,_that.month,_that.valueCents,_that.valueRate,_that.valueBool);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String key, @JsonKey(includeIfNull: false)  int? month, @JsonKey(includeIfNull: false, name: 'value_cents')  NonNegativeCents? valueCents, @JsonKey(includeIfNull: false, name: 'value_rate')  DecimalRate? valueRate, @JsonKey(includeIfNull: false, name: 'value_bool')  bool? valueBool)?  $default,) {final _that = this;
switch (_that) {
case _TaxInputValue() when $default != null:
return $default(_that.key,_that.month,_that.valueCents,_that.valueRate,_that.valueBool);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxInputValue implements TaxInputValue {
  const _TaxInputValue({required this.key, @JsonKey(includeIfNull: false) this.month, @JsonKey(includeIfNull: false, name: 'value_cents') this.valueCents, @JsonKey(includeIfNull: false, name: 'value_rate') this.valueRate, @JsonKey(includeIfNull: false, name: 'value_bool') this.valueBool});
  factory _TaxInputValue.fromJson(Map<String, dynamic> json) => _$TaxInputValueFromJson(json);

@override final  String key;
/// Month of the fiscal year, only for `money_monthly`.
@override@JsonKey(includeIfNull: false) final  int? month;
@override@JsonKey(includeIfNull: false, name: 'value_cents') final  NonNegativeCents? valueCents;
@override@JsonKey(includeIfNull: false, name: 'value_rate') final  DecimalRate? valueRate;
@override@JsonKey(includeIfNull: false, name: 'value_bool') final  bool? valueBool;

/// Create a copy of TaxInputValue
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxInputValueCopyWith<_TaxInputValue> get copyWith => __$TaxInputValueCopyWithImpl<_TaxInputValue>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxInputValueToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxInputValue&&(identical(other.key, key) || other.key == key)&&(identical(other.month, month) || other.month == month)&&(identical(other.valueCents, valueCents) || other.valueCents == valueCents)&&(identical(other.valueRate, valueRate) || other.valueRate == valueRate)&&(identical(other.valueBool, valueBool) || other.valueBool == valueBool));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,key,month,valueCents,valueRate,valueBool);
}

@override
String toString() {
    return 'TaxInputValue(key: $key, month: $month, valueCents: $valueCents, valueRate: $valueRate, valueBool: $valueBool)';
}


}

/// @nodoc
abstract mixin class _$TaxInputValueCopyWith<$Res> implements $TaxInputValueCopyWith<$Res> {
  factory _$TaxInputValueCopyWith(_TaxInputValue value, $Res Function(_TaxInputValue) _then) = __$TaxInputValueCopyWithImpl;
@override @useResult
$Res call({
 String key,@JsonKey(includeIfNull: false) int? month,@JsonKey(includeIfNull: false, name: 'value_cents') NonNegativeCents? valueCents,@JsonKey(includeIfNull: false, name: 'value_rate') DecimalRate? valueRate,@JsonKey(includeIfNull: false, name: 'value_bool') bool? valueBool
});




}
/// @nodoc
class __$TaxInputValueCopyWithImpl<$Res>
    implements _$TaxInputValueCopyWith<$Res> {
  __$TaxInputValueCopyWithImpl(this._self, this._then);

  final _TaxInputValue _self;
  final $Res Function(_TaxInputValue) _then;

/// Create a copy of TaxInputValue
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? key = null,Object? month = freezed,Object? valueCents = freezed,Object? valueRate = freezed,Object? valueBool = freezed,}) {
  return _then(_TaxInputValue(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,month: freezed == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int?,valueCents: freezed == valueCents ? _self.valueCents : valueCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,valueRate: freezed == valueRate ? _self.valueRate : valueRate // ignore: cast_nullable_to_non_nullable
as DecimalRate?,valueBool: freezed == valueBool ? _self.valueBool : valueBool // ignore: cast_nullable_to_non_nullable
as bool?,
  ));
}


}

// dart format on
