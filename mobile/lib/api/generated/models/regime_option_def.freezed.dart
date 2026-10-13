// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'regime_option_def.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RegimeOptionDef {

 String get key; String get label; RegimeOptionDefType get type;/// The name has been replaced because it contains a keyword. Original name: `required`.
@JsonKey(name: 'required') bool get requiredValue;@JsonKey(name: 'allowed_values') List<String>? get allowedValues;/// The name has been replaced because it contains a keyword. Original name: `default`.
/// The name has been replaced because it contains a keyword. Original name: `default`.
@JsonKey(name: 'default') dynamic get defaultValue;
/// Create a copy of RegimeOptionDef
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegimeOptionDefCopyWith<RegimeOptionDef> get copyWith => _$RegimeOptionDefCopyWithImpl<RegimeOptionDef>(this as RegimeOptionDef, _$identity);

  /// Serializes this RegimeOptionDef to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RegimeOptionDef;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegimeOptionDef&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.label, _this.label) || other.label == _this.label)&&(identical(other.type, _this.type) || other.type == _this.type)&&(identical(other.requiredValue, _this.requiredValue) || other.requiredValue == _this.requiredValue)&&const DeepCollectionEquality().equals(other.allowedValues, _this.allowedValues)&&const DeepCollectionEquality().equals(other.defaultValue, _this.defaultValue));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegimeOptionDef;
  return Object.hash(runtimeType,_this.key,_this.label,_this.type,_this.requiredValue,const DeepCollectionEquality().hash(_this.allowedValues),const DeepCollectionEquality().hash(_this.defaultValue));
}

@override
String toString() {
  final _this = this as RegimeOptionDef;
  return 'RegimeOptionDef(key: ${_this.key}, label: ${_this.label}, type: ${_this.type}, requiredValue: ${_this.requiredValue}, allowedValues: ${_this.allowedValues}, defaultValue: ${_this.defaultValue})';
}


}

/// @nodoc
abstract mixin class $RegimeOptionDefCopyWith<$Res>  {
  factory $RegimeOptionDefCopyWith(RegimeOptionDef value, $Res Function(RegimeOptionDef) _then) = _$RegimeOptionDefCopyWithImpl;
@useResult
$Res call({
 String key, String label, RegimeOptionDefType type,@JsonKey(name: 'required') bool requiredValue,@JsonKey(name: 'allowed_values') List<String>? allowedValues,@JsonKey(name: 'default') dynamic defaultValue
});




}
/// @nodoc
class _$RegimeOptionDefCopyWithImpl<$Res>
    implements $RegimeOptionDefCopyWith<$Res> {
  _$RegimeOptionDefCopyWithImpl(this._self, this._then);

  final RegimeOptionDef _self;
  final $Res Function(RegimeOptionDef) _then;

/// Create a copy of RegimeOptionDef
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? key = null,Object? label = null,Object? type = null,Object? requiredValue = null,Object? allowedValues = freezed,Object? defaultValue = freezed,}) {
  return _then(RegimeOptionDef(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as RegimeOptionDefType,requiredValue: null == requiredValue ? _self.requiredValue : requiredValue // ignore: cast_nullable_to_non_nullable
as bool,allowedValues: freezed == allowedValues ? _self.allowedValues : allowedValues // ignore: cast_nullable_to_non_nullable
as List<String>?,defaultValue: freezed == defaultValue ? _self.defaultValue : defaultValue // ignore: cast_nullable_to_non_nullable
as dynamic,
  ));
}

}


/// Adds pattern-matching-related methods to [RegimeOptionDef].
extension RegimeOptionDefPatterns on RegimeOptionDef {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RegimeOptionDef value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RegimeOptionDef() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RegimeOptionDef value)  $default,){
final _that = this;
switch (_that) {
case _RegimeOptionDef():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RegimeOptionDef value)?  $default,){
final _that = this;
switch (_that) {
case _RegimeOptionDef() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String key,  String label,  RegimeOptionDefType type, @JsonKey(name: 'required')  bool requiredValue, @JsonKey(name: 'allowed_values')  List<String>? allowedValues, @JsonKey(name: 'default')  dynamic defaultValue)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegimeOptionDef() when $default != null:
return $default(_that.key,_that.label,_that.type,_that.requiredValue,_that.allowedValues,_that.defaultValue);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String key,  String label,  RegimeOptionDefType type, @JsonKey(name: 'required')  bool requiredValue, @JsonKey(name: 'allowed_values')  List<String>? allowedValues, @JsonKey(name: 'default')  dynamic defaultValue)  $default,) {final _that = this;
switch (_that) {
case _RegimeOptionDef():
return $default(_that.key,_that.label,_that.type,_that.requiredValue,_that.allowedValues,_that.defaultValue);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String key,  String label,  RegimeOptionDefType type, @JsonKey(name: 'required')  bool requiredValue, @JsonKey(name: 'allowed_values')  List<String>? allowedValues, @JsonKey(name: 'default')  dynamic defaultValue)?  $default,) {final _that = this;
switch (_that) {
case _RegimeOptionDef() when $default != null:
return $default(_that.key,_that.label,_that.type,_that.requiredValue,_that.allowedValues,_that.defaultValue);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegimeOptionDef implements RegimeOptionDef {
  const _RegimeOptionDef({required this.key, required this.label, required this.type, @JsonKey(name: 'required') required this.requiredValue, @JsonKey(name: 'allowed_values')  List<String>? allowedValues, @JsonKey(name: 'default') this.defaultValue}): _allowedValues = allowedValues;
  factory _RegimeOptionDef.fromJson(Map<String, dynamic> json) => _$RegimeOptionDefFromJson(json);

@override final  String key;
@override final  String label;
@override final  RegimeOptionDefType type;
/// The name has been replaced because it contains a keyword. Original name: `required`.
@override@JsonKey(name: 'required') final  bool requiredValue;
 final  List<String>? _allowedValues;
@override@JsonKey(name: 'allowed_values') List<String>? get allowedValues {
  final value = _allowedValues;
  if (value == null) return null;
  if (_allowedValues is EqualUnmodifiableListView) return _allowedValues;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}

/// The name has been replaced because it contains a keyword. Original name: `default`.
/// The name has been replaced because it contains a keyword. Original name: `default`.
@override@JsonKey(name: 'default') final  dynamic defaultValue;

/// Create a copy of RegimeOptionDef
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegimeOptionDefCopyWith<_RegimeOptionDef> get copyWith => __$RegimeOptionDefCopyWithImpl<_RegimeOptionDef>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegimeOptionDefToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegimeOptionDef&&(identical(other.key, key) || other.key == key)&&(identical(other.label, label) || other.label == label)&&(identical(other.type, type) || other.type == type)&&(identical(other.requiredValue, requiredValue) || other.requiredValue == requiredValue)&&const DeepCollectionEquality().equals(other.allowedValues, _allowedValues)&&const DeepCollectionEquality().equals(other.defaultValue, defaultValue));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,key,label,type,requiredValue,const DeepCollectionEquality().hash(_allowedValues),const DeepCollectionEquality().hash(defaultValue));
}

@override
String toString() {
    return 'RegimeOptionDef(key: $key, label: $label, type: $type, requiredValue: $requiredValue, allowedValues: $allowedValues, defaultValue: $defaultValue)';
}


}

/// @nodoc
abstract mixin class _$RegimeOptionDefCopyWith<$Res> implements $RegimeOptionDefCopyWith<$Res> {
  factory _$RegimeOptionDefCopyWith(_RegimeOptionDef value, $Res Function(_RegimeOptionDef) _then) = __$RegimeOptionDefCopyWithImpl;
@override @useResult
$Res call({
 String key, String label, RegimeOptionDefType type,@JsonKey(name: 'required') bool requiredValue,@JsonKey(name: 'allowed_values') List<String>? allowedValues,@JsonKey(name: 'default') dynamic defaultValue
});




}
/// @nodoc
class __$RegimeOptionDefCopyWithImpl<$Res>
    implements _$RegimeOptionDefCopyWith<$Res> {
  __$RegimeOptionDefCopyWithImpl(this._self, this._then);

  final _RegimeOptionDef _self;
  final $Res Function(_RegimeOptionDef) _then;

/// Create a copy of RegimeOptionDef
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? key = null,Object? label = null,Object? type = null,Object? requiredValue = null,Object? allowedValues = freezed,Object? defaultValue = freezed,}) {
  return _then(_RegimeOptionDef(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as RegimeOptionDefType,requiredValue: null == requiredValue ? _self.requiredValue : requiredValue // ignore: cast_nullable_to_non_nullable
as bool,allowedValues: freezed == allowedValues ? _self._allowedValues : allowedValues // ignore: cast_nullable_to_non_nullable
as List<String>?,defaultValue: freezed == defaultValue ? _self.defaultValue : defaultValue // ignore: cast_nullable_to_non_nullable
as dynamic,
  ));
}


}

// dart format on
