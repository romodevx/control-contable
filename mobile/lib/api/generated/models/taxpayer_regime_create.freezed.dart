// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_regime_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerRegimeCreate {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(includeIfNull: false) RegimeOptions? get options;@JsonKey(includeIfNull: false, name: 'start_period') Period? get startPeriod;
/// Create a copy of TaxpayerRegimeCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerRegimeCreateCopyWith<TaxpayerRegimeCreate> get copyWith => _$TaxpayerRegimeCreateCopyWithImpl<TaxpayerRegimeCreate>(this as TaxpayerRegimeCreate, _$identity);

  /// Serializes this TaxpayerRegimeCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerRegimeCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerRegimeCreate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&const DeepCollectionEquality().equals(other.options, _this.options)&&(identical(other.startPeriod, _this.startPeriod) || other.startPeriod == _this.startPeriod));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerRegimeCreate;
  return Object.hash(runtimeType,_this.regimeCode,const DeepCollectionEquality().hash(_this.options),_this.startPeriod);
}

@override
String toString() {
  final _this = this as TaxpayerRegimeCreate;
  return 'TaxpayerRegimeCreate(regimeCode: ${_this.regimeCode}, options: ${_this.options}, startPeriod: ${_this.startPeriod})';
}


}

/// @nodoc
abstract mixin class $TaxpayerRegimeCreateCopyWith<$Res>  {
  factory $TaxpayerRegimeCreateCopyWith(TaxpayerRegimeCreate value, $Res Function(TaxpayerRegimeCreate) _then) = _$TaxpayerRegimeCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(includeIfNull: false) RegimeOptions? options,@JsonKey(includeIfNull: false, name: 'start_period') Period? startPeriod
});




}
/// @nodoc
class _$TaxpayerRegimeCreateCopyWithImpl<$Res>
    implements $TaxpayerRegimeCreateCopyWith<$Res> {
  _$TaxpayerRegimeCreateCopyWithImpl(this._self, this._then);

  final TaxpayerRegimeCreate _self;
  final $Res Function(TaxpayerRegimeCreate) _then;

/// Create a copy of TaxpayerRegimeCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? options = freezed,Object? startPeriod = freezed,}) {
  return _then(TaxpayerRegimeCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,options: freezed == options ? _self.options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions?,startPeriod: freezed == startPeriod ? _self.startPeriod : startPeriod // ignore: cast_nullable_to_non_nullable
as Period?,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerRegimeCreate].
extension TaxpayerRegimeCreatePatterns on TaxpayerRegimeCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerRegimeCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerRegimeCreate value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerRegimeCreate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(includeIfNull: false)  RegimeOptions? options, @JsonKey(includeIfNull: false, name: 'start_period')  Period? startPeriod)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate() when $default != null:
return $default(_that.regimeCode,_that.options,_that.startPeriod);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(includeIfNull: false)  RegimeOptions? options, @JsonKey(includeIfNull: false, name: 'start_period')  Period? startPeriod)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate():
return $default(_that.regimeCode,_that.options,_that.startPeriod);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(includeIfNull: false)  RegimeOptions? options, @JsonKey(includeIfNull: false, name: 'start_period')  Period? startPeriod)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegimeCreate() when $default != null:
return $default(_that.regimeCode,_that.options,_that.startPeriod);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerRegimeCreate implements TaxpayerRegimeCreate {
  const _TaxpayerRegimeCreate({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(includeIfNull: false)  RegimeOptions? options, @JsonKey(includeIfNull: false, name: 'start_period') this.startPeriod}): _options = options;
  factory _TaxpayerRegimeCreate.fromJson(Map<String, dynamic> json) => _$TaxpayerRegimeCreateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
 final  RegimeOptions? _options;
@override@JsonKey(includeIfNull: false) RegimeOptions? get options {
  final value = _options;
  if (value == null) return null;
  if (_options is EqualUnmodifiableMapView) return _options;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(value);
}

@override@JsonKey(includeIfNull: false, name: 'start_period') final  Period? startPeriod;

/// Create a copy of TaxpayerRegimeCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerRegimeCreateCopyWith<_TaxpayerRegimeCreate> get copyWith => __$TaxpayerRegimeCreateCopyWithImpl<_TaxpayerRegimeCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerRegimeCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerRegimeCreate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&const DeepCollectionEquality().equals(other.options, _options)&&(identical(other.startPeriod, startPeriod) || other.startPeriod == startPeriod));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,const DeepCollectionEquality().hash(_options),startPeriod);
}

@override
String toString() {
    return 'TaxpayerRegimeCreate(regimeCode: $regimeCode, options: $options, startPeriod: $startPeriod)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerRegimeCreateCopyWith<$Res> implements $TaxpayerRegimeCreateCopyWith<$Res> {
  factory _$TaxpayerRegimeCreateCopyWith(_TaxpayerRegimeCreate value, $Res Function(_TaxpayerRegimeCreate) _then) = __$TaxpayerRegimeCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(includeIfNull: false) RegimeOptions? options,@JsonKey(includeIfNull: false, name: 'start_period') Period? startPeriod
});




}
/// @nodoc
class __$TaxpayerRegimeCreateCopyWithImpl<$Res>
    implements _$TaxpayerRegimeCreateCopyWith<$Res> {
  __$TaxpayerRegimeCreateCopyWithImpl(this._self, this._then);

  final _TaxpayerRegimeCreate _self;
  final $Res Function(_TaxpayerRegimeCreate) _then;

/// Create a copy of TaxpayerRegimeCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? options = freezed,Object? startPeriod = freezed,}) {
  return _then(_TaxpayerRegimeCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,options: freezed == options ? _self._options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions?,startPeriod: freezed == startPeriod ? _self.startPeriod : startPeriod // ignore: cast_nullable_to_non_nullable
as Period?,
  ));
}


}

// dart format on
