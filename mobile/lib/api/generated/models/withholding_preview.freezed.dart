// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'withholding_preview.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$WithholdingPreview {

 Withholding get withholding; List<Warning> get warnings;
/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$WithholdingPreviewCopyWith<WithholdingPreview> get copyWith => _$WithholdingPreviewCopyWithImpl<WithholdingPreview>(this as WithholdingPreview, _$identity);

  /// Serializes this WithholdingPreview to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as WithholdingPreview;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is WithholdingPreview&&(identical(other.withholding, _this.withholding) || other.withholding == _this.withholding)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as WithholdingPreview;
  return Object.hash(runtimeType,_this.withholding,const DeepCollectionEquality().hash(_this.warnings));
}

@override
String toString() {
  final _this = this as WithholdingPreview;
  return 'WithholdingPreview(withholding: ${_this.withholding}, warnings: ${_this.warnings})';
}


}

/// @nodoc
abstract mixin class $WithholdingPreviewCopyWith<$Res>  {
  factory $WithholdingPreviewCopyWith(WithholdingPreview value, $Res Function(WithholdingPreview) _then) = _$WithholdingPreviewCopyWithImpl;
@useResult
$Res call({
 Withholding withholding, List<Warning> warnings
});


$WithholdingCopyWith<$Res> get withholding;

}
/// @nodoc
class _$WithholdingPreviewCopyWithImpl<$Res>
    implements $WithholdingPreviewCopyWith<$Res> {
  _$WithholdingPreviewCopyWithImpl(this._self, this._then);

  final WithholdingPreview _self;
  final $Res Function(WithholdingPreview) _then;

/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? withholding = null,Object? warnings = null,}) {
  return _then(WithholdingPreview(
withholding: null == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,
  ));
}
/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res> get withholding {
  
  return $WithholdingCopyWith<$Res>(_self.withholding, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}


/// Adds pattern-matching-related methods to [WithholdingPreview].
extension WithholdingPreviewPatterns on WithholdingPreview {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _WithholdingPreview value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _WithholdingPreview() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _WithholdingPreview value)  $default,){
final _that = this;
switch (_that) {
case _WithholdingPreview():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _WithholdingPreview value)?  $default,){
final _that = this;
switch (_that) {
case _WithholdingPreview() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Withholding withholding,  List<Warning> warnings)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _WithholdingPreview() when $default != null:
return $default(_that.withholding,_that.warnings);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Withholding withholding,  List<Warning> warnings)  $default,) {final _that = this;
switch (_that) {
case _WithholdingPreview():
return $default(_that.withholding,_that.warnings);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Withholding withholding,  List<Warning> warnings)?  $default,) {final _that = this;
switch (_that) {
case _WithholdingPreview() when $default != null:
return $default(_that.withholding,_that.warnings);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _WithholdingPreview implements WithholdingPreview {
  const _WithholdingPreview({required this.withholding, required  List<Warning> warnings}): _warnings = warnings;
  factory _WithholdingPreview.fromJson(Map<String, dynamic> json) => _$WithholdingPreviewFromJson(json);

@override final  Withholding withholding;
 final  List<Warning> _warnings;
@override List<Warning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}


/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$WithholdingPreviewCopyWith<_WithholdingPreview> get copyWith => __$WithholdingPreviewCopyWithImpl<_WithholdingPreview>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$WithholdingPreviewToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _WithholdingPreview&&(identical(other.withholding, withholding) || other.withholding == withholding)&&const DeepCollectionEquality().equals(other.warnings, _warnings));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,withholding,const DeepCollectionEquality().hash(_warnings));
}

@override
String toString() {
    return 'WithholdingPreview(withholding: $withholding, warnings: $warnings)';
}


}

/// @nodoc
abstract mixin class _$WithholdingPreviewCopyWith<$Res> implements $WithholdingPreviewCopyWith<$Res> {
  factory _$WithholdingPreviewCopyWith(_WithholdingPreview value, $Res Function(_WithholdingPreview) _then) = __$WithholdingPreviewCopyWithImpl;
@override @useResult
$Res call({
 Withholding withholding, List<Warning> warnings
});


@override $WithholdingCopyWith<$Res> get withholding;

}
/// @nodoc
class __$WithholdingPreviewCopyWithImpl<$Res>
    implements _$WithholdingPreviewCopyWith<$Res> {
  __$WithholdingPreviewCopyWithImpl(this._self, this._then);

  final _WithholdingPreview _self;
  final $Res Function(_WithholdingPreview) _then;

/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? withholding = null,Object? warnings = null,}) {
  return _then(_WithholdingPreview(
withholding: null == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,
  ));
}

/// Create a copy of WithholdingPreview
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res> get withholding {
  
  return $WithholdingCopyWith<$Res>(_self.withholding, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}

// dart format on
