// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_input_def.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxInputDef {

 String get key;/// Spanish label shown in the tax-inputs form.
 String get label; TaxInputDefKind get kind;/// The name has been replaced because it contains a keyword. Original name: `required`.
@JsonKey(name: 'required') bool get requiredValue;
/// Create a copy of TaxInputDef
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxInputDefCopyWith<TaxInputDef> get copyWith => _$TaxInputDefCopyWithImpl<TaxInputDef>(this as TaxInputDef, _$identity);

  /// Serializes this TaxInputDef to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxInputDef;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxInputDef&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.label, _this.label) || other.label == _this.label)&&(identical(other.kind, _this.kind) || other.kind == _this.kind)&&(identical(other.requiredValue, _this.requiredValue) || other.requiredValue == _this.requiredValue));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxInputDef;
  return Object.hash(runtimeType,_this.key,_this.label,_this.kind,_this.requiredValue);
}

@override
String toString() {
  final _this = this as TaxInputDef;
  return 'TaxInputDef(key: ${_this.key}, label: ${_this.label}, kind: ${_this.kind}, requiredValue: ${_this.requiredValue})';
}


}

/// @nodoc
abstract mixin class $TaxInputDefCopyWith<$Res>  {
  factory $TaxInputDefCopyWith(TaxInputDef value, $Res Function(TaxInputDef) _then) = _$TaxInputDefCopyWithImpl;
@useResult
$Res call({
 String key, String label, TaxInputDefKind kind,@JsonKey(name: 'required') bool requiredValue
});




}
/// @nodoc
class _$TaxInputDefCopyWithImpl<$Res>
    implements $TaxInputDefCopyWith<$Res> {
  _$TaxInputDefCopyWithImpl(this._self, this._then);

  final TaxInputDef _self;
  final $Res Function(TaxInputDef) _then;

/// Create a copy of TaxInputDef
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? key = null,Object? label = null,Object? kind = null,Object? requiredValue = null,}) {
  return _then(TaxInputDef(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as TaxInputDefKind,requiredValue: null == requiredValue ? _self.requiredValue : requiredValue // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxInputDef].
extension TaxInputDefPatterns on TaxInputDef {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxInputDef value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxInputDef() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxInputDef value)  $default,){
final _that = this;
switch (_that) {
case _TaxInputDef():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxInputDef value)?  $default,){
final _that = this;
switch (_that) {
case _TaxInputDef() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String key,  String label,  TaxInputDefKind kind, @JsonKey(name: 'required')  bool requiredValue)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxInputDef() when $default != null:
return $default(_that.key,_that.label,_that.kind,_that.requiredValue);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String key,  String label,  TaxInputDefKind kind, @JsonKey(name: 'required')  bool requiredValue)  $default,) {final _that = this;
switch (_that) {
case _TaxInputDef():
return $default(_that.key,_that.label,_that.kind,_that.requiredValue);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String key,  String label,  TaxInputDefKind kind, @JsonKey(name: 'required')  bool requiredValue)?  $default,) {final _that = this;
switch (_that) {
case _TaxInputDef() when $default != null:
return $default(_that.key,_that.label,_that.kind,_that.requiredValue);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxInputDef implements TaxInputDef {
  const _TaxInputDef({required this.key, required this.label, required this.kind, @JsonKey(name: 'required') required this.requiredValue});
  factory _TaxInputDef.fromJson(Map<String, dynamic> json) => _$TaxInputDefFromJson(json);

@override final  String key;
/// Spanish label shown in the tax-inputs form.
@override final  String label;
@override final  TaxInputDefKind kind;
/// The name has been replaced because it contains a keyword. Original name: `required`.
@override@JsonKey(name: 'required') final  bool requiredValue;

/// Create a copy of TaxInputDef
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxInputDefCopyWith<_TaxInputDef> get copyWith => __$TaxInputDefCopyWithImpl<_TaxInputDef>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxInputDefToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxInputDef&&(identical(other.key, key) || other.key == key)&&(identical(other.label, label) || other.label == label)&&(identical(other.kind, kind) || other.kind == kind)&&(identical(other.requiredValue, requiredValue) || other.requiredValue == requiredValue));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,key,label,kind,requiredValue);
}

@override
String toString() {
    return 'TaxInputDef(key: $key, label: $label, kind: $kind, requiredValue: $requiredValue)';
}


}

/// @nodoc
abstract mixin class _$TaxInputDefCopyWith<$Res> implements $TaxInputDefCopyWith<$Res> {
  factory _$TaxInputDefCopyWith(_TaxInputDef value, $Res Function(_TaxInputDef) _then) = __$TaxInputDefCopyWithImpl;
@override @useResult
$Res call({
 String key, String label, TaxInputDefKind kind,@JsonKey(name: 'required') bool requiredValue
});




}
/// @nodoc
class __$TaxInputDefCopyWithImpl<$Res>
    implements _$TaxInputDefCopyWith<$Res> {
  __$TaxInputDefCopyWithImpl(this._self, this._then);

  final _TaxInputDef _self;
  final $Res Function(_TaxInputDef) _then;

/// Create a copy of TaxInputDef
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? key = null,Object? label = null,Object? kind = null,Object? requiredValue = null,}) {
  return _then(_TaxInputDef(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as TaxInputDefKind,requiredValue: null == requiredValue ? _self.requiredValue : requiredValue // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
