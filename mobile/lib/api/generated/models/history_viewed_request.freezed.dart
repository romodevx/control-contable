// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_viewed_request.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryViewedRequest {

@JsonKey(name: 'records_fingerprint') RecordsFingerprint get recordsFingerprint;@JsonKey(name: 'figures_fingerprint') FiguresFingerprint get figuresFingerprint;@JsonKey(name: 'engine_version') EngineVersion get engineVersion;@JsonKey(name: 'parameter_set') HistoryViewedParameterSet get parameterSet;
/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryViewedRequestCopyWith<HistoryViewedRequest> get copyWith => _$HistoryViewedRequestCopyWithImpl<HistoryViewedRequest>(this as HistoryViewedRequest, _$identity);

  /// Serializes this HistoryViewedRequest to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryViewedRequest;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryViewedRequest&&(identical(other.recordsFingerprint, _this.recordsFingerprint) || other.recordsFingerprint == _this.recordsFingerprint)&&(identical(other.figuresFingerprint, _this.figuresFingerprint) || other.figuresFingerprint == _this.figuresFingerprint)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryViewedRequest;
  return Object.hash(runtimeType,_this.recordsFingerprint,_this.figuresFingerprint,_this.engineVersion,_this.parameterSet);
}

@override
String toString() {
  final _this = this as HistoryViewedRequest;
  return 'HistoryViewedRequest(recordsFingerprint: ${_this.recordsFingerprint}, figuresFingerprint: ${_this.figuresFingerprint}, engineVersion: ${_this.engineVersion}, parameterSet: ${_this.parameterSet})';
}


}

/// @nodoc
abstract mixin class $HistoryViewedRequestCopyWith<$Res>  {
  factory $HistoryViewedRequestCopyWith(HistoryViewedRequest value, $Res Function(HistoryViewedRequest) _then) = _$HistoryViewedRequestCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'figures_fingerprint') FiguresFingerprint figuresFingerprint,@JsonKey(name: 'engine_version') EngineVersion engineVersion,@JsonKey(name: 'parameter_set') HistoryViewedParameterSet parameterSet
});


$HistoryViewedParameterSetCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$HistoryViewedRequestCopyWithImpl<$Res>
    implements $HistoryViewedRequestCopyWith<$Res> {
  _$HistoryViewedRequestCopyWithImpl(this._self, this._then);

  final HistoryViewedRequest _self;
  final $Res Function(HistoryViewedRequest) _then;

/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? recordsFingerprint = null,Object? figuresFingerprint = null,Object? engineVersion = null,Object? parameterSet = null,}) {
  return _then(HistoryViewedRequest(
recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,figuresFingerprint: null == figuresFingerprint ? _self.figuresFingerprint : figuresFingerprint // ignore: cast_nullable_to_non_nullable
as FiguresFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as HistoryViewedParameterSet,
  ));
}
/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryViewedParameterSetCopyWith<$Res> get parameterSet {
  
  return $HistoryViewedParameterSetCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}


/// Adds pattern-matching-related methods to [HistoryViewedRequest].
extension HistoryViewedRequestPatterns on HistoryViewedRequest {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryViewedRequest value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryViewedRequest() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryViewedRequest value)  $default,){
final _that = this;
switch (_that) {
case _HistoryViewedRequest():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryViewedRequest value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryViewedRequest() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  HistoryViewedParameterSet parameterSet)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryViewedRequest() when $default != null:
return $default(_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  HistoryViewedParameterSet parameterSet)  $default,) {final _that = this;
switch (_that) {
case _HistoryViewedRequest():
return $default(_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  HistoryViewedParameterSet parameterSet)?  $default,) {final _that = this;
switch (_that) {
case _HistoryViewedRequest() when $default != null:
return $default(_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryViewedRequest implements HistoryViewedRequest {
  const _HistoryViewedRequest({@JsonKey(name: 'records_fingerprint') required this.recordsFingerprint, @JsonKey(name: 'figures_fingerprint') required this.figuresFingerprint, @JsonKey(name: 'engine_version') required this.engineVersion, @JsonKey(name: 'parameter_set') required this.parameterSet});
  factory _HistoryViewedRequest.fromJson(Map<String, dynamic> json) => _$HistoryViewedRequestFromJson(json);

@override@JsonKey(name: 'records_fingerprint') final  RecordsFingerprint recordsFingerprint;
@override@JsonKey(name: 'figures_fingerprint') final  FiguresFingerprint figuresFingerprint;
@override@JsonKey(name: 'engine_version') final  EngineVersion engineVersion;
@override@JsonKey(name: 'parameter_set') final  HistoryViewedParameterSet parameterSet;

/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryViewedRequestCopyWith<_HistoryViewedRequest> get copyWith => __$HistoryViewedRequestCopyWithImpl<_HistoryViewedRequest>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryViewedRequestToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryViewedRequest&&(identical(other.recordsFingerprint, recordsFingerprint) || other.recordsFingerprint == recordsFingerprint)&&(identical(other.figuresFingerprint, figuresFingerprint) || other.figuresFingerprint == figuresFingerprint)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,recordsFingerprint,figuresFingerprint,engineVersion,parameterSet);
}

@override
String toString() {
    return 'HistoryViewedRequest(recordsFingerprint: $recordsFingerprint, figuresFingerprint: $figuresFingerprint, engineVersion: $engineVersion, parameterSet: $parameterSet)';
}


}

/// @nodoc
abstract mixin class _$HistoryViewedRequestCopyWith<$Res> implements $HistoryViewedRequestCopyWith<$Res> {
  factory _$HistoryViewedRequestCopyWith(_HistoryViewedRequest value, $Res Function(_HistoryViewedRequest) _then) = __$HistoryViewedRequestCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'figures_fingerprint') FiguresFingerprint figuresFingerprint,@JsonKey(name: 'engine_version') EngineVersion engineVersion,@JsonKey(name: 'parameter_set') HistoryViewedParameterSet parameterSet
});


@override $HistoryViewedParameterSetCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$HistoryViewedRequestCopyWithImpl<$Res>
    implements _$HistoryViewedRequestCopyWith<$Res> {
  __$HistoryViewedRequestCopyWithImpl(this._self, this._then);

  final _HistoryViewedRequest _self;
  final $Res Function(_HistoryViewedRequest) _then;

/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? recordsFingerprint = null,Object? figuresFingerprint = null,Object? engineVersion = null,Object? parameterSet = null,}) {
  return _then(_HistoryViewedRequest(
recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,figuresFingerprint: null == figuresFingerprint ? _self.figuresFingerprint : figuresFingerprint // ignore: cast_nullable_to_non_nullable
as FiguresFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as HistoryViewedParameterSet,
  ));
}

/// Create a copy of HistoryViewedRequest
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryViewedParameterSetCopyWith<$Res> get parameterSet {
  
  return $HistoryViewedParameterSetCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}

// dart format on
