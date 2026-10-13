// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'document_acceptance.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DocumentAcceptance {

 LegalDocumentKind get kind;/// Opaque version label (data, not code).
 String get version;
/// Create a copy of DocumentAcceptance
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DocumentAcceptanceCopyWith<DocumentAcceptance> get copyWith => _$DocumentAcceptanceCopyWithImpl<DocumentAcceptance>(this as DocumentAcceptance, _$identity);

  /// Serializes this DocumentAcceptance to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DocumentAcceptance;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DocumentAcceptance&&(identical(other.kind, _this.kind) || other.kind == _this.kind)&&(identical(other.version, _this.version) || other.version == _this.version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DocumentAcceptance;
  return Object.hash(runtimeType,_this.kind,_this.version);
}

@override
String toString() {
  final _this = this as DocumentAcceptance;
  return 'DocumentAcceptance(kind: ${_this.kind}, version: ${_this.version})';
}


}

/// @nodoc
abstract mixin class $DocumentAcceptanceCopyWith<$Res>  {
  factory $DocumentAcceptanceCopyWith(DocumentAcceptance value, $Res Function(DocumentAcceptance) _then) = _$DocumentAcceptanceCopyWithImpl;
@useResult
$Res call({
 LegalDocumentKind kind, String version
});




}
/// @nodoc
class _$DocumentAcceptanceCopyWithImpl<$Res>
    implements $DocumentAcceptanceCopyWith<$Res> {
  _$DocumentAcceptanceCopyWithImpl(this._self, this._then);

  final DocumentAcceptance _self;
  final $Res Function(DocumentAcceptance) _then;

/// Create a copy of DocumentAcceptance
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? kind = null,Object? version = null,}) {
  return _then(DocumentAcceptance(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as LegalDocumentKind,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [DocumentAcceptance].
extension DocumentAcceptancePatterns on DocumentAcceptance {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DocumentAcceptance value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DocumentAcceptance() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DocumentAcceptance value)  $default,){
final _that = this;
switch (_that) {
case _DocumentAcceptance():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DocumentAcceptance value)?  $default,){
final _that = this;
switch (_that) {
case _DocumentAcceptance() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( LegalDocumentKind kind,  String version)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DocumentAcceptance() when $default != null:
return $default(_that.kind,_that.version);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( LegalDocumentKind kind,  String version)  $default,) {final _that = this;
switch (_that) {
case _DocumentAcceptance():
return $default(_that.kind,_that.version);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( LegalDocumentKind kind,  String version)?  $default,) {final _that = this;
switch (_that) {
case _DocumentAcceptance() when $default != null:
return $default(_that.kind,_that.version);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DocumentAcceptance implements DocumentAcceptance {
  const _DocumentAcceptance({required this.kind, required this.version});
  factory _DocumentAcceptance.fromJson(Map<String, dynamic> json) => _$DocumentAcceptanceFromJson(json);

@override final  LegalDocumentKind kind;
/// Opaque version label (data, not code).
@override final  String version;

/// Create a copy of DocumentAcceptance
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DocumentAcceptanceCopyWith<_DocumentAcceptance> get copyWith => __$DocumentAcceptanceCopyWithImpl<_DocumentAcceptance>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DocumentAcceptanceToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DocumentAcceptance&&(identical(other.kind, kind) || other.kind == kind)&&(identical(other.version, version) || other.version == version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,kind,version);
}

@override
String toString() {
    return 'DocumentAcceptance(kind: $kind, version: $version)';
}


}

/// @nodoc
abstract mixin class _$DocumentAcceptanceCopyWith<$Res> implements $DocumentAcceptanceCopyWith<$Res> {
  factory _$DocumentAcceptanceCopyWith(_DocumentAcceptance value, $Res Function(_DocumentAcceptance) _then) = __$DocumentAcceptanceCopyWithImpl;
@override @useResult
$Res call({
 LegalDocumentKind kind, String version
});




}
/// @nodoc
class __$DocumentAcceptanceCopyWithImpl<$Res>
    implements _$DocumentAcceptanceCopyWith<$Res> {
  __$DocumentAcceptanceCopyWithImpl(this._self, this._then);

  final _DocumentAcceptance _self;
  final $Res Function(_DocumentAcceptance) _then;

/// Create a copy of DocumentAcceptance
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? kind = null,Object? version = null,}) {
  return _then(_DocumentAcceptance(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as LegalDocumentKind,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
