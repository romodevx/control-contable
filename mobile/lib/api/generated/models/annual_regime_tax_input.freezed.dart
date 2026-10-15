// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_regime_tax_input.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualRegimeTaxInput {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode; String get key; String get label;@JsonKey(includeIfNull: true, name: 'value_cents') NonNegativeCents? get valueCents;
/// Create a copy of AnnualRegimeTaxInput
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualRegimeTaxInputCopyWith<AnnualRegimeTaxInput> get copyWith => _$AnnualRegimeTaxInputCopyWithImpl<AnnualRegimeTaxInput>(this as AnnualRegimeTaxInput, _$identity);

  /// Serializes this AnnualRegimeTaxInput to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualRegimeTaxInput;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualRegimeTaxInput&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.label, _this.label) || other.label == _this.label)&&(identical(other.valueCents, _this.valueCents) || other.valueCents == _this.valueCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualRegimeTaxInput;
  return Object.hash(runtimeType,_this.regimeCode,_this.key,_this.label,_this.valueCents);
}

@override
String toString() {
  final _this = this as AnnualRegimeTaxInput;
  return 'AnnualRegimeTaxInput(regimeCode: ${_this.regimeCode}, key: ${_this.key}, label: ${_this.label}, valueCents: ${_this.valueCents})';
}


}

/// @nodoc
abstract mixin class $AnnualRegimeTaxInputCopyWith<$Res>  {
  factory $AnnualRegimeTaxInputCopyWith(AnnualRegimeTaxInput value, $Res Function(AnnualRegimeTaxInput) _then) = _$AnnualRegimeTaxInputCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, String key, String label,@JsonKey(includeIfNull: true, name: 'value_cents') NonNegativeCents? valueCents
});




}
/// @nodoc
class _$AnnualRegimeTaxInputCopyWithImpl<$Res>
    implements $AnnualRegimeTaxInputCopyWith<$Res> {
  _$AnnualRegimeTaxInputCopyWithImpl(this._self, this._then);

  final AnnualRegimeTaxInput _self;
  final $Res Function(AnnualRegimeTaxInput) _then;

/// Create a copy of AnnualRegimeTaxInput
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? key = null,Object? label = null,Object? valueCents = freezed,}) {
  return _then(AnnualRegimeTaxInput(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,valueCents: freezed == valueCents ? _self.valueCents : valueCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualRegimeTaxInput].
extension AnnualRegimeTaxInputPatterns on AnnualRegimeTaxInput {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualRegimeTaxInput value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualRegimeTaxInput value)  $default,){
final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualRegimeTaxInput value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String key,  String label, @JsonKey(includeIfNull: true, name: 'value_cents')  NonNegativeCents? valueCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput() when $default != null:
return $default(_that.regimeCode,_that.key,_that.label,_that.valueCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String key,  String label, @JsonKey(includeIfNull: true, name: 'value_cents')  NonNegativeCents? valueCents)  $default,) {final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput():
return $default(_that.regimeCode,_that.key,_that.label,_that.valueCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  String key,  String label, @JsonKey(includeIfNull: true, name: 'value_cents')  NonNegativeCents? valueCents)?  $default,) {final _that = this;
switch (_that) {
case _AnnualRegimeTaxInput() when $default != null:
return $default(_that.regimeCode,_that.key,_that.label,_that.valueCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualRegimeTaxInput implements AnnualRegimeTaxInput {
  const _AnnualRegimeTaxInput({@JsonKey(name: 'regime_code') required this.regimeCode, required this.key, required this.label, @JsonKey(includeIfNull: true, name: 'value_cents') required this.valueCents});
  factory _AnnualRegimeTaxInput.fromJson(Map<String, dynamic> json) => _$AnnualRegimeTaxInputFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override final  String key;
@override final  String label;
@override@JsonKey(includeIfNull: true, name: 'value_cents') final  NonNegativeCents? valueCents;

/// Create a copy of AnnualRegimeTaxInput
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualRegimeTaxInputCopyWith<_AnnualRegimeTaxInput> get copyWith => __$AnnualRegimeTaxInputCopyWithImpl<_AnnualRegimeTaxInput>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualRegimeTaxInputToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualRegimeTaxInput&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.key, key) || other.key == key)&&(identical(other.label, label) || other.label == label)&&(identical(other.valueCents, valueCents) || other.valueCents == valueCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,key,label,valueCents);
}

@override
String toString() {
    return 'AnnualRegimeTaxInput(regimeCode: $regimeCode, key: $key, label: $label, valueCents: $valueCents)';
}


}

/// @nodoc
abstract mixin class _$AnnualRegimeTaxInputCopyWith<$Res> implements $AnnualRegimeTaxInputCopyWith<$Res> {
  factory _$AnnualRegimeTaxInputCopyWith(_AnnualRegimeTaxInput value, $Res Function(_AnnualRegimeTaxInput) _then) = __$AnnualRegimeTaxInputCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, String key, String label,@JsonKey(includeIfNull: true, name: 'value_cents') NonNegativeCents? valueCents
});




}
/// @nodoc
class __$AnnualRegimeTaxInputCopyWithImpl<$Res>
    implements _$AnnualRegimeTaxInputCopyWith<$Res> {
  __$AnnualRegimeTaxInputCopyWithImpl(this._self, this._then);

  final _AnnualRegimeTaxInput _self;
  final $Res Function(_AnnualRegimeTaxInput) _then;

/// Create a copy of AnnualRegimeTaxInput
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? key = null,Object? label = null,Object? valueCents = freezed,}) {
  return _then(_AnnualRegimeTaxInput(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,valueCents: freezed == valueCents ? _self.valueCents : valueCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}


}

// dart format on
