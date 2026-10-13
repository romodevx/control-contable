// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'breakdown.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Breakdown {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId; Period get period; RegimeEstimate get regime; List<BreakdownLine> get lines; Legend get legend; List<Disclaimer> get disclaimers;@JsonKey(name: 'engine_version') String get engineVersion;
/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$BreakdownCopyWith<Breakdown> get copyWith => _$BreakdownCopyWithImpl<Breakdown>(this as Breakdown, _$identity);

  /// Serializes this Breakdown to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Breakdown;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Breakdown&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.regime, _this.regime) || other.regime == _this.regime)&&const DeepCollectionEquality().equals(other.lines, _this.lines)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&const DeepCollectionEquality().equals(other.disclaimers, _this.disclaimers)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Breakdown;
  return Object.hash(runtimeType,_this.taxpayerId,_this.period,_this.regime,const DeepCollectionEquality().hash(_this.lines),_this.legend,const DeepCollectionEquality().hash(_this.disclaimers),_this.engineVersion);
}

@override
String toString() {
  final _this = this as Breakdown;
  return 'Breakdown(taxpayerId: ${_this.taxpayerId}, period: ${_this.period}, regime: ${_this.regime}, lines: ${_this.lines}, legend: ${_this.legend}, disclaimers: ${_this.disclaimers}, engineVersion: ${_this.engineVersion})';
}


}

/// @nodoc
abstract mixin class $BreakdownCopyWith<$Res>  {
  factory $BreakdownCopyWith(Breakdown value, $Res Function(Breakdown) _then) = _$BreakdownCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period, RegimeEstimate regime, List<BreakdownLine> lines, Legend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'engine_version') String engineVersion
});


$RegimeEstimateCopyWith<$Res> get regime;

}
/// @nodoc
class _$BreakdownCopyWithImpl<$Res>
    implements $BreakdownCopyWith<$Res> {
  _$BreakdownCopyWithImpl(this._self, this._then);

  final Breakdown _self;
  final $Res Function(Breakdown) _then;

/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? period = null,Object? regime = null,Object? lines = null,Object? legend = null,Object? disclaimers = null,Object? engineVersion = null,}) {
  return _then(Breakdown(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,regime: null == regime ? _self.regime : regime // ignore: cast_nullable_to_non_nullable
as RegimeEstimate,lines: null == lines ? _self.lines : lines // ignore: cast_nullable_to_non_nullable
as List<BreakdownLine>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self.disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as String,
  ));
}
/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RegimeEstimateCopyWith<$Res> get regime {
  
  return $RegimeEstimateCopyWith<$Res>(_self.regime, (value) {
    return _then(_self.copyWith(regime: value));
  });
}
}


/// Adds pattern-matching-related methods to [Breakdown].
extension BreakdownPatterns on Breakdown {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Breakdown value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Breakdown() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Breakdown value)  $default,){
final _that = this;
switch (_that) {
case _Breakdown():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Breakdown value)?  $default,){
final _that = this;
switch (_that) {
case _Breakdown() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period,  RegimeEstimate regime,  List<BreakdownLine> lines,  Legend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  String engineVersion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Breakdown() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.regime,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period,  RegimeEstimate regime,  List<BreakdownLine> lines,  Legend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  String engineVersion)  $default,) {final _that = this;
switch (_that) {
case _Breakdown():
return $default(_that.taxpayerId,_that.period,_that.regime,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period,  RegimeEstimate regime,  List<BreakdownLine> lines,  Legend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  String engineVersion)?  $default,) {final _that = this;
switch (_that) {
case _Breakdown() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.regime,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Breakdown implements Breakdown {
  const _Breakdown({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, required this.period, required this.regime, required  List<BreakdownLine> lines, required this.legend, required  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version') required this.engineVersion}): _lines = lines,_disclaimers = disclaimers;
  factory _Breakdown.fromJson(Map<String, dynamic> json) => _$BreakdownFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override final  Period period;
@override final  RegimeEstimate regime;
 final  List<BreakdownLine> _lines;
@override List<BreakdownLine> get lines {
  if (_lines is EqualUnmodifiableListView) return _lines;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_lines);
}

@override final  Legend legend;
 final  List<Disclaimer> _disclaimers;
@override List<Disclaimer> get disclaimers {
  if (_disclaimers is EqualUnmodifiableListView) return _disclaimers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_disclaimers);
}

@override@JsonKey(name: 'engine_version') final  String engineVersion;

/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$BreakdownCopyWith<_Breakdown> get copyWith => __$BreakdownCopyWithImpl<_Breakdown>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$BreakdownToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Breakdown&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.period, period) || other.period == period)&&(identical(other.regime, regime) || other.regime == regime)&&const DeepCollectionEquality().equals(other.lines, _lines)&&(identical(other.legend, legend) || other.legend == legend)&&const DeepCollectionEquality().equals(other.disclaimers, _disclaimers)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,period,regime,const DeepCollectionEquality().hash(_lines),legend,const DeepCollectionEquality().hash(_disclaimers),engineVersion);
}

@override
String toString() {
    return 'Breakdown(taxpayerId: $taxpayerId, period: $period, regime: $regime, lines: $lines, legend: $legend, disclaimers: $disclaimers, engineVersion: $engineVersion)';
}


}

/// @nodoc
abstract mixin class _$BreakdownCopyWith<$Res> implements $BreakdownCopyWith<$Res> {
  factory _$BreakdownCopyWith(_Breakdown value, $Res Function(_Breakdown) _then) = __$BreakdownCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period, RegimeEstimate regime, List<BreakdownLine> lines, Legend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'engine_version') String engineVersion
});


@override $RegimeEstimateCopyWith<$Res> get regime;

}
/// @nodoc
class __$BreakdownCopyWithImpl<$Res>
    implements _$BreakdownCopyWith<$Res> {
  __$BreakdownCopyWithImpl(this._self, this._then);

  final _Breakdown _self;
  final $Res Function(_Breakdown) _then;

/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? period = null,Object? regime = null,Object? lines = null,Object? legend = null,Object? disclaimers = null,Object? engineVersion = null,}) {
  return _then(_Breakdown(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,regime: null == regime ? _self.regime : regime // ignore: cast_nullable_to_non_nullable
as RegimeEstimate,lines: null == lines ? _self._lines : lines // ignore: cast_nullable_to_non_nullable
as List<BreakdownLine>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self._disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

/// Create a copy of Breakdown
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$RegimeEstimateCopyWith<$Res> get regime {
  
  return $RegimeEstimateCopyWith<$Res>(_self.regime, (value) {
    return _then(_self.copyWith(regime: value));
  });
}
}

// dart format on
