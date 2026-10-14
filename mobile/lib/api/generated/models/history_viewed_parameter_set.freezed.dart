// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_viewed_parameter_set.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryViewedParameterSet {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; int get version;
/// Create a copy of HistoryViewedParameterSet
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryViewedParameterSetCopyWith<HistoryViewedParameterSet> get copyWith => _$HistoryViewedParameterSetCopyWithImpl<HistoryViewedParameterSet>(this as HistoryViewedParameterSet, _$identity);

  /// Serializes this HistoryViewedParameterSet to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryViewedParameterSet;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryViewedParameterSet&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.version, _this.version) || other.version == _this.version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryViewedParameterSet;
  return Object.hash(runtimeType,_this.fiscalYear,_this.version);
}

@override
String toString() {
  final _this = this as HistoryViewedParameterSet;
  return 'HistoryViewedParameterSet(fiscalYear: ${_this.fiscalYear}, version: ${_this.version})';
}


}

/// @nodoc
abstract mixin class $HistoryViewedParameterSetCopyWith<$Res>  {
  factory $HistoryViewedParameterSetCopyWith(HistoryViewedParameterSet value, $Res Function(HistoryViewedParameterSet) _then) = _$HistoryViewedParameterSetCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, int version
});




}
/// @nodoc
class _$HistoryViewedParameterSetCopyWithImpl<$Res>
    implements $HistoryViewedParameterSetCopyWith<$Res> {
  _$HistoryViewedParameterSetCopyWithImpl(this._self, this._then);

  final HistoryViewedParameterSet _self;
  final $Res Function(HistoryViewedParameterSet) _then;

/// Create a copy of HistoryViewedParameterSet
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? version = null,}) {
  return _then(HistoryViewedParameterSet(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [HistoryViewedParameterSet].
extension HistoryViewedParameterSetPatterns on HistoryViewedParameterSet {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryViewedParameterSet value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryViewedParameterSet() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryViewedParameterSet value)  $default,){
final _that = this;
switch (_that) {
case _HistoryViewedParameterSet():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryViewedParameterSet value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryViewedParameterSet() when $default != null:
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
case _HistoryViewedParameterSet() when $default != null:
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
case _HistoryViewedParameterSet():
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
case _HistoryViewedParameterSet() when $default != null:
return $default(_that.fiscalYear,_that.version);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryViewedParameterSet implements HistoryViewedParameterSet {
  const _HistoryViewedParameterSet({@JsonKey(name: 'fiscal_year') required this.fiscalYear, required this.version});
  factory _HistoryViewedParameterSet.fromJson(Map<String, dynamic> json) => _$HistoryViewedParameterSetFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override final  int version;

/// Create a copy of HistoryViewedParameterSet
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryViewedParameterSetCopyWith<_HistoryViewedParameterSet> get copyWith => __$HistoryViewedParameterSetCopyWithImpl<_HistoryViewedParameterSet>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryViewedParameterSetToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryViewedParameterSet&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.version, version) || other.version == version));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,version);
}

@override
String toString() {
    return 'HistoryViewedParameterSet(fiscalYear: $fiscalYear, version: $version)';
}


}

/// @nodoc
abstract mixin class _$HistoryViewedParameterSetCopyWith<$Res> implements $HistoryViewedParameterSetCopyWith<$Res> {
  factory _$HistoryViewedParameterSetCopyWith(_HistoryViewedParameterSet value, $Res Function(_HistoryViewedParameterSet) _then) = __$HistoryViewedParameterSetCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, int version
});




}
/// @nodoc
class __$HistoryViewedParameterSetCopyWithImpl<$Res>
    implements _$HistoryViewedParameterSetCopyWith<$Res> {
  __$HistoryViewedParameterSetCopyWithImpl(this._self, this._then);

  final _HistoryViewedParameterSet _self;
  final $Res Function(_HistoryViewedParameterSet) _then;

/// Create a copy of HistoryViewedParameterSet
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? version = null,}) {
  return _then(_HistoryViewedParameterSet(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,version: null == version ? _self.version : version // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
