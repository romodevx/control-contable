// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'warning.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Warning {

 String get code;/// Same dotted naming as `FieldError.field`.
 String get field;/// Safe to show (es-MX); never contains the RFC.
 String get message;
/// Create a copy of Warning
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$WarningCopyWith<Warning> get copyWith => _$WarningCopyWithImpl<Warning>(this as Warning, _$identity);

  /// Serializes this Warning to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Warning;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Warning&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.field, _this.field) || other.field == _this.field)&&(identical(other.message, _this.message) || other.message == _this.message));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Warning;
  return Object.hash(runtimeType,_this.code,_this.field,_this.message);
}

@override
String toString() {
  final _this = this as Warning;
  return 'Warning(code: ${_this.code}, field: ${_this.field}, message: ${_this.message})';
}


}

/// @nodoc
abstract mixin class $WarningCopyWith<$Res>  {
  factory $WarningCopyWith(Warning value, $Res Function(Warning) _then) = _$WarningCopyWithImpl;
@useResult
$Res call({
 String code, String field, String message
});




}
/// @nodoc
class _$WarningCopyWithImpl<$Res>
    implements $WarningCopyWith<$Res> {
  _$WarningCopyWithImpl(this._self, this._then);

  final Warning _self;
  final $Res Function(Warning) _then;

/// Create a copy of Warning
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? field = null,Object? message = null,}) {
  return _then(Warning(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,field: null == field ? _self.field : field // ignore: cast_nullable_to_non_nullable
as String,message: null == message ? _self.message : message // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [Warning].
extension WarningPatterns on Warning {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Warning value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Warning() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Warning value)  $default,){
final _that = this;
switch (_that) {
case _Warning():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Warning value)?  $default,){
final _that = this;
switch (_that) {
case _Warning() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String code,  String field,  String message)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Warning() when $default != null:
return $default(_that.code,_that.field,_that.message);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String code,  String field,  String message)  $default,) {final _that = this;
switch (_that) {
case _Warning():
return $default(_that.code,_that.field,_that.message);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String code,  String field,  String message)?  $default,) {final _that = this;
switch (_that) {
case _Warning() when $default != null:
return $default(_that.code,_that.field,_that.message);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Warning implements Warning {
  const _Warning({required this.code, required this.field, required this.message});
  factory _Warning.fromJson(Map<String, dynamic> json) => _$WarningFromJson(json);

@override final  String code;
/// Same dotted naming as `FieldError.field`.
@override final  String field;
/// Safe to show (es-MX); never contains the RFC.
@override final  String message;

/// Create a copy of Warning
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$WarningCopyWith<_Warning> get copyWith => __$WarningCopyWithImpl<_Warning>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$WarningToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Warning&&(identical(other.code, code) || other.code == code)&&(identical(other.field, field) || other.field == field)&&(identical(other.message, message) || other.message == message));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,field,message);
}

@override
String toString() {
    return 'Warning(code: $code, field: $field, message: $message)';
}


}

/// @nodoc
abstract mixin class _$WarningCopyWith<$Res> implements $WarningCopyWith<$Res> {
  factory _$WarningCopyWith(_Warning value, $Res Function(_Warning) _then) = __$WarningCopyWithImpl;
@override @useResult
$Res call({
 String code, String field, String message
});




}
/// @nodoc
class __$WarningCopyWithImpl<$Res>
    implements _$WarningCopyWith<$Res> {
  __$WarningCopyWithImpl(this._self, this._then);

  final _Warning _self;
  final $Res Function(_Warning) _then;

/// Create a copy of Warning
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? field = null,Object? message = null,}) {
  return _then(_Warning(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,field: null == field ? _self.field : field // ignore: cast_nullable_to_non_nullable
as String,message: null == message ? _self.message : message // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
