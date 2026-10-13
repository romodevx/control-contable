// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'parameter_set_ref.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ParameterSetRef {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; int get version;
/// Create a copy of ParameterSetRef
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<ParameterSetRef> get copyWith => _$ParameterSetRefCopyWithImpl<ParameterSetRef>(this as ParameterSetRef, _$identity);

  /// Serializes this ParameterSetRef to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ParameterSetRef;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ParameterSetRef&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.version, _this.version) || other.version == _this.version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ParameterSetRef;
  return Object.hash(runtimeType,_this.fiscalYear,_this.version);
}

@override
String toString() {
  final _this = this as ParameterSetRef;
  return 'ParameterSetRef(fiscalYear: ${_this.fiscalYear}, version: ${_this.version})';
}


}

/// @nodoc
abstract mixin class $ParameterSetRefCopyWith<$Res>  {
  factory $ParameterSetRefCopyWith(ParameterSetRef value, $Res Function(ParameterSetRef) _then) = _$ParameterSetRefCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, int version
});




}
/// @nodoc
class _$ParameterSetRefCopyWithImpl<$Res>
    implements $ParameterSetRefCopyWith<$Res> {
  _$ParameterSetRefCopyWithImpl(this._self, this._then);

  final ParameterSetRef _self;
  final $Res Function(ParameterSetRef) _then;

/// Create a copy of ParameterSetRef
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? version = null,}) {
  return _then(ParameterSetRef(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [ParameterSetRef].
extension ParameterSetRefPatterns on ParameterSetRef {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ParameterSetRef value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ParameterSetRef() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ParameterSetRef value)  $default,){
final _that = this;
switch (_that) {
case _ParameterSetRef():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ParameterSetRef value)?  $default,){
final _that = this;
switch (_that) {
case _ParameterSetRef() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  int version)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ParameterSetRef() when $default != null:
return $default(_that.fiscalYear,_that.version);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  int version)  $default,) {final _that = this;
switch (_that) {
case _ParameterSetRef():
return $default(_that.fiscalYear,_that.version);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  int version)?  $default,) {final _that = this;
switch (_that) {
case _ParameterSetRef() when $default != null:
return $default(_that.fiscalYear,_that.version);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ParameterSetRef implements ParameterSetRef {
  const _ParameterSetRef({@JsonKey(name: 'fiscal_year') required this.fiscalYear, required this.version});
  factory _ParameterSetRef.fromJson(Map<String, dynamic> json) => _$ParameterSetRefFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override final  int version;

/// Create a copy of ParameterSetRef
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ParameterSetRefCopyWith<_ParameterSetRef> get copyWith => __$ParameterSetRefCopyWithImpl<_ParameterSetRef>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ParameterSetRefToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ParameterSetRef&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.version, version) || other.version == version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,version);
}

@override
String toString() {
    return 'ParameterSetRef(fiscalYear: $fiscalYear, version: $version)';
}


}

/// @nodoc
abstract mixin class _$ParameterSetRefCopyWith<$Res> implements $ParameterSetRefCopyWith<$Res> {
  factory _$ParameterSetRefCopyWith(_ParameterSetRef value, $Res Function(_ParameterSetRef) _then) = __$ParameterSetRefCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, int version
});




}
/// @nodoc
class __$ParameterSetRefCopyWithImpl<$Res>
    implements _$ParameterSetRefCopyWith<$Res> {
  __$ParameterSetRefCopyWithImpl(this._self, this._then);

  final _ParameterSetRef _self;
  final $Res Function(_ParameterSetRef) _then;

/// Create a copy of ParameterSetRef
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? version = null,}) {
  return _then(_ParameterSetRef(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
