// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_warning.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseWarning {

 ExpenseWarningCode get code; String get message;
/// Create a copy of ExpenseWarning
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseWarningCopyWith<ExpenseWarning> get copyWith => _$ExpenseWarningCopyWithImpl<ExpenseWarning>(this as ExpenseWarning, _$identity);

  /// Serializes this ExpenseWarning to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseWarning;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseWarning&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.message, _this.message) || other.message == _this.message));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseWarning;
  return Object.hash(runtimeType,_this.code,_this.message);
}

@override
String toString() {
  final _this = this as ExpenseWarning;
  return 'ExpenseWarning(code: ${_this.code}, message: ${_this.message})';
}


}

/// @nodoc
abstract mixin class $ExpenseWarningCopyWith<$Res>  {
  factory $ExpenseWarningCopyWith(ExpenseWarning value, $Res Function(ExpenseWarning) _then) = _$ExpenseWarningCopyWithImpl;
@useResult
$Res call({
 ExpenseWarningCode code, String message
});




}
/// @nodoc
class _$ExpenseWarningCopyWithImpl<$Res>
    implements $ExpenseWarningCopyWith<$Res> {
  _$ExpenseWarningCopyWithImpl(this._self, this._then);

  final ExpenseWarning _self;
  final $Res Function(ExpenseWarning) _then;

/// Create a copy of ExpenseWarning
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? message = null,}) {
  return _then(ExpenseWarning(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as ExpenseWarningCode,message: null == message ? _self.message : message // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseWarning].
extension ExpenseWarningPatterns on ExpenseWarning {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseWarning value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseWarning() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseWarning value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseWarning():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseWarning value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseWarning() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( ExpenseWarningCode code,  String message)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseWarning() when $default != null:
return $default(_that.code,_that.message);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( ExpenseWarningCode code,  String message)  $default,) {final _that = this;
switch (_that) {
case _ExpenseWarning():
return $default(_that.code,_that.message);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( ExpenseWarningCode code,  String message)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseWarning() when $default != null:
return $default(_that.code,_that.message);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseWarning implements ExpenseWarning {
  const _ExpenseWarning({required this.code, required this.message});
  factory _ExpenseWarning.fromJson(Map<String, dynamic> json) => _$ExpenseWarningFromJson(json);

@override final  ExpenseWarningCode code;
@override final  String message;

/// Create a copy of ExpenseWarning
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseWarningCopyWith<_ExpenseWarning> get copyWith => __$ExpenseWarningCopyWithImpl<_ExpenseWarning>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseWarningToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseWarning&&(identical(other.code, code) || other.code == code)&&(identical(other.message, message) || other.message == message));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,message);
}

@override
String toString() {
    return 'ExpenseWarning(code: $code, message: $message)';
}


}

/// @nodoc
abstract mixin class _$ExpenseWarningCopyWith<$Res> implements $ExpenseWarningCopyWith<$Res> {
  factory _$ExpenseWarningCopyWith(_ExpenseWarning value, $Res Function(_ExpenseWarning) _then) = __$ExpenseWarningCopyWithImpl;
@override @useResult
$Res call({
 ExpenseWarningCode code, String message
});




}
/// @nodoc
class __$ExpenseWarningCopyWithImpl<$Res>
    implements _$ExpenseWarningCopyWith<$Res> {
  __$ExpenseWarningCopyWithImpl(this._self, this._then);

  final _ExpenseWarning _self;
  final $Res Function(_ExpenseWarning) _then;

/// Create a copy of ExpenseWarning
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? message = null,}) {
  return _then(_ExpenseWarning(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as ExpenseWarningCode,message: null == message ? _self.message : message // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
