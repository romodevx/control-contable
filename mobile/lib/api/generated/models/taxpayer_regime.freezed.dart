// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_regime.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerRegime {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode; String get name; Treatment get treatment; RegimeOptions get options;@JsonKey(name: 'start_period') Period get startPeriod;@JsonKey(includeIfNull: true, name: 'end_period') Period? get endPeriod;
/// Create a copy of TaxpayerRegime
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerRegimeCopyWith<TaxpayerRegime> get copyWith => _$TaxpayerRegimeCopyWithImpl<TaxpayerRegime>(this as TaxpayerRegime, _$identity);

  /// Serializes this TaxpayerRegime to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerRegime;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerRegime&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.name, _this.name) || other.name == _this.name)&&(identical(other.treatment, _this.treatment) || other.treatment == _this.treatment)&&const DeepCollectionEquality().equals(other.options, _this.options)&&(identical(other.startPeriod, _this.startPeriod) || other.startPeriod == _this.startPeriod)&&(identical(other.endPeriod, _this.endPeriod) || other.endPeriod == _this.endPeriod));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerRegime;
  return Object.hash(runtimeType,_this.regimeCode,_this.name,_this.treatment,const DeepCollectionEquality().hash(_this.options),_this.startPeriod,_this.endPeriod);
}

@override
String toString() {
  final _this = this as TaxpayerRegime;
  return 'TaxpayerRegime(regimeCode: ${_this.regimeCode}, name: ${_this.name}, treatment: ${_this.treatment}, options: ${_this.options}, startPeriod: ${_this.startPeriod}, endPeriod: ${_this.endPeriod})';
}


}

/// @nodoc
abstract mixin class $TaxpayerRegimeCopyWith<$Res>  {
  factory $TaxpayerRegimeCopyWith(TaxpayerRegime value, $Res Function(TaxpayerRegime) _then) = _$TaxpayerRegimeCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, String name, Treatment treatment, RegimeOptions options,@JsonKey(name: 'start_period') Period startPeriod,@JsonKey(includeIfNull: true, name: 'end_period') Period? endPeriod
});




}
/// @nodoc
class _$TaxpayerRegimeCopyWithImpl<$Res>
    implements $TaxpayerRegimeCopyWith<$Res> {
  _$TaxpayerRegimeCopyWithImpl(this._self, this._then);

  final TaxpayerRegime _self;
  final $Res Function(TaxpayerRegime) _then;

/// Create a copy of TaxpayerRegime
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? name = null,Object? treatment = null,Object? options = null,Object? startPeriod = null,Object? endPeriod = freezed,}) {
  return _then(TaxpayerRegime(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,options: null == options ? _self.options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions,startPeriod: null == startPeriod ? _self.startPeriod : startPeriod // ignore: cast_nullable_to_non_nullable
as Period,endPeriod: freezed == endPeriod ? _self.endPeriod : endPeriod // ignore: cast_nullable_to_non_nullable
as Period?,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerRegime].
extension TaxpayerRegimePatterns on TaxpayerRegime {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerRegime value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerRegime() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerRegime value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegime():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerRegime value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegime() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String name,  Treatment treatment,  RegimeOptions options, @JsonKey(name: 'start_period')  Period startPeriod, @JsonKey(includeIfNull: true, name: 'end_period')  Period? endPeriod)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerRegime() when $default != null:
return $default(_that.regimeCode,_that.name,_that.treatment,_that.options,_that.startPeriod,_that.endPeriod);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String name,  Treatment treatment,  RegimeOptions options, @JsonKey(name: 'start_period')  Period startPeriod, @JsonKey(includeIfNull: true, name: 'end_period')  Period? endPeriod)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegime():
return $default(_that.regimeCode,_that.name,_that.treatment,_that.options,_that.startPeriod,_that.endPeriod);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String name,  Treatment treatment,  RegimeOptions options, @JsonKey(name: 'start_period')  Period startPeriod, @JsonKey(includeIfNull: true, name: 'end_period')  Period? endPeriod)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegime() when $default != null:
return $default(_that.regimeCode,_that.name,_that.treatment,_that.options,_that.startPeriod,_that.endPeriod);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerRegime implements TaxpayerRegime {
  const _TaxpayerRegime({@JsonKey(name: 'regime_code') required this.regimeCode, required this.name, required this.treatment, required  RegimeOptions options, @JsonKey(name: 'start_period') required this.startPeriod, @JsonKey(includeIfNull: true, name: 'end_period') required this.endPeriod}): _options = options;
  factory _TaxpayerRegime.fromJson(Map<String, dynamic> json) => _$TaxpayerRegimeFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override final  String name;
@override final  Treatment treatment;
 final  RegimeOptions _options;
@override RegimeOptions get options {
  if (_options is EqualUnmodifiableMapView) return _options;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(_options);
}

@override@JsonKey(name: 'start_period') final  Period startPeriod;
@override@JsonKey(includeIfNull: true, name: 'end_period') final  Period? endPeriod;

/// Create a copy of TaxpayerRegime
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerRegimeCopyWith<_TaxpayerRegime> get copyWith => __$TaxpayerRegimeCopyWithImpl<_TaxpayerRegime>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerRegimeToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerRegime&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.name, name) || other.name == name)&&(identical(other.treatment, treatment) || other.treatment == treatment)&&const DeepCollectionEquality().equals(other.options, _options)&&(identical(other.startPeriod, startPeriod) || other.startPeriod == startPeriod)&&(identical(other.endPeriod, endPeriod) || other.endPeriod == endPeriod));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,name,treatment,const DeepCollectionEquality().hash(_options),startPeriod,endPeriod);
}

@override
String toString() {
    return 'TaxpayerRegime(regimeCode: $regimeCode, name: $name, treatment: $treatment, options: $options, startPeriod: $startPeriod, endPeriod: $endPeriod)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerRegimeCopyWith<$Res> implements $TaxpayerRegimeCopyWith<$Res> {
  factory _$TaxpayerRegimeCopyWith(_TaxpayerRegime value, $Res Function(_TaxpayerRegime) _then) = __$TaxpayerRegimeCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, String name, Treatment treatment, RegimeOptions options,@JsonKey(name: 'start_period') Period startPeriod,@JsonKey(includeIfNull: true, name: 'end_period') Period? endPeriod
});




}
/// @nodoc
class __$TaxpayerRegimeCopyWithImpl<$Res>
    implements _$TaxpayerRegimeCopyWith<$Res> {
  __$TaxpayerRegimeCopyWithImpl(this._self, this._then);

  final _TaxpayerRegime _self;
  final $Res Function(_TaxpayerRegime) _then;

/// Create a copy of TaxpayerRegime
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? name = null,Object? treatment = null,Object? options = null,Object? startPeriod = null,Object? endPeriod = freezed,}) {
  return _then(_TaxpayerRegime(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,options: null == options ? _self._options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions,startPeriod: null == startPeriod ? _self.startPeriod : startPeriod // ignore: cast_nullable_to_non_nullable
as Period,endPeriod: freezed == endPeriod ? _self.endPeriod : endPeriod // ignore: cast_nullable_to_non_nullable
as Period?,
  ));
}


}

// dart format on
