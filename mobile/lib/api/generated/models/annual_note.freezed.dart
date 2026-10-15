// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_note.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualNote {

 AnnualNoteCode get code; String get text;
/// Create a copy of AnnualNote
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualNoteCopyWith<AnnualNote> get copyWith => _$AnnualNoteCopyWithImpl<AnnualNote>(this as AnnualNote, _$identity);

  /// Serializes this AnnualNote to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualNote;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualNote&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.text, _this.text) || other.text == _this.text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualNote;
  return Object.hash(runtimeType,_this.code,_this.text);
}

@override
String toString() {
  final _this = this as AnnualNote;
  return 'AnnualNote(code: ${_this.code}, text: ${_this.text})';
}


}

/// @nodoc
abstract mixin class $AnnualNoteCopyWith<$Res>  {
  factory $AnnualNoteCopyWith(AnnualNote value, $Res Function(AnnualNote) _then) = _$AnnualNoteCopyWithImpl;
@useResult
$Res call({
 AnnualNoteCode code, String text
});




}
/// @nodoc
class _$AnnualNoteCopyWithImpl<$Res>
    implements $AnnualNoteCopyWith<$Res> {
  _$AnnualNoteCopyWithImpl(this._self, this._then);

  final AnnualNote _self;
  final $Res Function(AnnualNote) _then;

/// Create a copy of AnnualNote
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? text = null,}) {
  return _then(AnnualNote(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as AnnualNoteCode,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualNote].
extension AnnualNotePatterns on AnnualNote {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualNote value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualNote() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualNote value)  $default,){
final _that = this;
switch (_that) {
case _AnnualNote():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualNote value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualNote() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( AnnualNoteCode code,  String text)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualNote() when $default != null:
return $default(_that.code,_that.text);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( AnnualNoteCode code,  String text)  $default,) {final _that = this;
switch (_that) {
case _AnnualNote():
return $default(_that.code,_that.text);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( AnnualNoteCode code,  String text)?  $default,) {final _that = this;
switch (_that) {
case _AnnualNote() when $default != null:
return $default(_that.code,_that.text);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualNote implements AnnualNote {
  const _AnnualNote({required this.code, required this.text});
  factory _AnnualNote.fromJson(Map<String, dynamic> json) => _$AnnualNoteFromJson(json);

@override final  AnnualNoteCode code;
@override final  String text;

/// Create a copy of AnnualNote
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualNoteCopyWith<_AnnualNote> get copyWith => __$AnnualNoteCopyWithImpl<_AnnualNote>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualNoteToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualNote&&(identical(other.code, code) || other.code == code)&&(identical(other.text, text) || other.text == text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,text);
}

@override
String toString() {
    return 'AnnualNote(code: $code, text: $text)';
}


}

/// @nodoc
abstract mixin class _$AnnualNoteCopyWith<$Res> implements $AnnualNoteCopyWith<$Res> {
  factory _$AnnualNoteCopyWith(_AnnualNote value, $Res Function(_AnnualNote) _then) = __$AnnualNoteCopyWithImpl;
@override @useResult
$Res call({
 AnnualNoteCode code, String text
});




}
/// @nodoc
class __$AnnualNoteCopyWithImpl<$Res>
    implements _$AnnualNoteCopyWith<$Res> {
  __$AnnualNoteCopyWithImpl(this._self, this._then);

  final _AnnualNote _self;
  final $Res Function(_AnnualNote) _then;

/// Create a copy of AnnualNote
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? text = null,}) {
  return _then(_AnnualNote(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as AnnualNoteCode,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
