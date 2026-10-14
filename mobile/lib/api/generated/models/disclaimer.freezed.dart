// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'disclaimer.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Disclaimer {

 DisclaimerCode get code; DisclaimerSeverity get severity; String get text;
/// Create a copy of Disclaimer
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DisclaimerCopyWith<Disclaimer> get copyWith => _$DisclaimerCopyWithImpl<Disclaimer>(this as Disclaimer, _$identity);

  /// Serializes this Disclaimer to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Disclaimer;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Disclaimer&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.severity, _this.severity) || other.severity == _this.severity)&&(identical(other.text, _this.text) || other.text == _this.text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Disclaimer;
  return Object.hash(runtimeType,_this.code,_this.severity,_this.text);
}

@override
String toString() {
  final _this = this as Disclaimer;
  return 'Disclaimer(code: ${_this.code}, severity: ${_this.severity}, text: ${_this.text})';
}


}

/// @nodoc
abstract mixin class $DisclaimerCopyWith<$Res>  {
  factory $DisclaimerCopyWith(Disclaimer value, $Res Function(Disclaimer) _then) = _$DisclaimerCopyWithImpl;
@useResult
$Res call({
 DisclaimerCode code, DisclaimerSeverity severity, String text
});




}
/// @nodoc
class _$DisclaimerCopyWithImpl<$Res>
    implements $DisclaimerCopyWith<$Res> {
  _$DisclaimerCopyWithImpl(this._self, this._then);

  final Disclaimer _self;
  final $Res Function(Disclaimer) _then;

/// Create a copy of Disclaimer
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? severity = null,Object? text = null,}) {
  return _then(Disclaimer(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as DisclaimerCode,severity: null == severity ? _self.severity : severity // ignore: cast_nullable_to_non_nullable
as DisclaimerSeverity,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [Disclaimer].
extension DisclaimerPatterns on Disclaimer {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Disclaimer value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Disclaimer() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Disclaimer value)  $default,){
final _that = this;
switch (_that) {
case _Disclaimer():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Disclaimer value)?  $default,){
final _that = this;
switch (_that) {
case _Disclaimer() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( DisclaimerCode code,  DisclaimerSeverity severity,  String text)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Disclaimer() when $default != null:
return $default(_that.code,_that.severity,_that.text);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( DisclaimerCode code,  DisclaimerSeverity severity,  String text)  $default,) {final _that = this;
switch (_that) {
case _Disclaimer():
return $default(_that.code,_that.severity,_that.text);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( DisclaimerCode code,  DisclaimerSeverity severity,  String text)?  $default,) {final _that = this;
switch (_that) {
case _Disclaimer() when $default != null:
return $default(_that.code,_that.severity,_that.text);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Disclaimer implements Disclaimer {
  const _Disclaimer({required this.code, required this.severity, required this.text});
  factory _Disclaimer.fromJson(Map<String, dynamic> json) => _$DisclaimerFromJson(json);

@override final  DisclaimerCode code;
@override final  DisclaimerSeverity severity;
@override final  String text;

/// Create a copy of Disclaimer
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DisclaimerCopyWith<_Disclaimer> get copyWith => __$DisclaimerCopyWithImpl<_Disclaimer>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DisclaimerToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Disclaimer&&(identical(other.code, code) || other.code == code)&&(identical(other.severity, severity) || other.severity == severity)&&(identical(other.text, text) || other.text == text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,severity,text);
}

@override
String toString() {
    return 'Disclaimer(code: $code, severity: $severity, text: $text)';
}


}

/// @nodoc
abstract mixin class _$DisclaimerCopyWith<$Res> implements $DisclaimerCopyWith<$Res> {
  factory _$DisclaimerCopyWith(_Disclaimer value, $Res Function(_Disclaimer) _then) = __$DisclaimerCopyWithImpl;
@override @useResult
$Res call({
 DisclaimerCode code, DisclaimerSeverity severity, String text
});




}
/// @nodoc
class __$DisclaimerCopyWithImpl<$Res>
    implements _$DisclaimerCopyWith<$Res> {
  __$DisclaimerCopyWithImpl(this._self, this._then);

  final _Disclaimer _self;
  final $Res Function(_Disclaimer) _then;

/// Create a copy of Disclaimer
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? severity = null,Object? text = null,}) {
  return _then(_Disclaimer(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as DisclaimerCode,severity: null == severity ? _self.severity : severity // ignore: cast_nullable_to_non_nullable
as DisclaimerSeverity,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
