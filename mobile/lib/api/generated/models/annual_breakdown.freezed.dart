// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_breakdown.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualBreakdown {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; AnnualSection get section; List<AnnualBreakdownLine> get lines; AnnualLegend get legend; List<Disclaimer> get disclaimers;@JsonKey(name: 'engine_version') EngineVersion get engineVersion;
/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualBreakdownCopyWith<AnnualBreakdown> get copyWith => _$AnnualBreakdownCopyWithImpl<AnnualBreakdown>(this as AnnualBreakdown, _$identity);

  /// Serializes this AnnualBreakdown to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualBreakdown;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualBreakdown&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.section, _this.section) || other.section == _this.section)&&const DeepCollectionEquality().equals(other.lines, _this.lines)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&const DeepCollectionEquality().equals(other.disclaimers, _this.disclaimers)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualBreakdown;
  return Object.hash(runtimeType,_this.taxpayerId,_this.fiscalYear,_this.section,const DeepCollectionEquality().hash(_this.lines),_this.legend,const DeepCollectionEquality().hash(_this.disclaimers),_this.engineVersion);
}

@override
String toString() {
  final _this = this as AnnualBreakdown;
  return 'AnnualBreakdown(taxpayerId: ${_this.taxpayerId}, fiscalYear: ${_this.fiscalYear}, section: ${_this.section}, lines: ${_this.lines}, legend: ${_this.legend}, disclaimers: ${_this.disclaimers}, engineVersion: ${_this.engineVersion})';
}


}

/// @nodoc
abstract mixin class $AnnualBreakdownCopyWith<$Res>  {
  factory $AnnualBreakdownCopyWith(AnnualBreakdown value, $Res Function(AnnualBreakdown) _then) = _$AnnualBreakdownCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, AnnualSection section, List<AnnualBreakdownLine> lines, AnnualLegend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'engine_version') EngineVersion engineVersion
});


$AnnualSectionCopyWith<$Res> get section;

}
/// @nodoc
class _$AnnualBreakdownCopyWithImpl<$Res>
    implements $AnnualBreakdownCopyWith<$Res> {
  _$AnnualBreakdownCopyWithImpl(this._self, this._then);

  final AnnualBreakdown _self;
  final $Res Function(AnnualBreakdown) _then;

/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? section = null,Object? lines = null,Object? legend = null,Object? disclaimers = null,Object? engineVersion = null,}) {
  return _then(AnnualBreakdown(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,section: null == section ? _self.section : section // ignore: cast_nullable_to_non_nullable
as AnnualSection,lines: null == lines ? _self.lines : lines // ignore: cast_nullable_to_non_nullable
as List<AnnualBreakdownLine>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as AnnualLegend,disclaimers: null == disclaimers ? _self.disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,
  ));
}
/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualSectionCopyWith<$Res> get section {
  
  return $AnnualSectionCopyWith<$Res>(_self.section, (value) {
    return _then(_self.copyWith(section: value));
  });
}
}


/// Adds pattern-matching-related methods to [AnnualBreakdown].
extension AnnualBreakdownPatterns on AnnualBreakdown {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualBreakdown value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualBreakdown() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualBreakdown value)  $default,){
final _that = this;
switch (_that) {
case _AnnualBreakdown():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualBreakdown value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualBreakdown() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  AnnualSection section,  List<AnnualBreakdownLine> lines,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualBreakdown() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.section,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  AnnualSection section,  List<AnnualBreakdownLine> lines,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)  $default,) {final _that = this;
switch (_that) {
case _AnnualBreakdown():
return $default(_that.taxpayerId,_that.fiscalYear,_that.section,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  AnnualSection section,  List<AnnualBreakdownLine> lines,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)?  $default,) {final _that = this;
switch (_that) {
case _AnnualBreakdown() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.section,_that.lines,_that.legend,_that.disclaimers,_that.engineVersion);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualBreakdown implements AnnualBreakdown {
  const _AnnualBreakdown({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'fiscal_year') required this.fiscalYear, required this.section, required  List<AnnualBreakdownLine> lines, required this.legend, required  List<Disclaimer> disclaimers, @JsonKey(name: 'engine_version') required this.engineVersion}): _lines = lines,_disclaimers = disclaimers;
  factory _AnnualBreakdown.fromJson(Map<String, dynamic> json) => _$AnnualBreakdownFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override final  AnnualSection section;
 final  List<AnnualBreakdownLine> _lines;
@override List<AnnualBreakdownLine> get lines {
  if (_lines is EqualUnmodifiableListView) return _lines;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_lines);
}

@override final  AnnualLegend legend;
 final  List<Disclaimer> _disclaimers;
@override List<Disclaimer> get disclaimers {
  if (_disclaimers is EqualUnmodifiableListView) return _disclaimers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_disclaimers);
}

@override@JsonKey(name: 'engine_version') final  EngineVersion engineVersion;

/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualBreakdownCopyWith<_AnnualBreakdown> get copyWith => __$AnnualBreakdownCopyWithImpl<_AnnualBreakdown>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualBreakdownToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualBreakdown&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.section, section) || other.section == section)&&const DeepCollectionEquality().equals(other.lines, _lines)&&(identical(other.legend, legend) || other.legend == legend)&&const DeepCollectionEquality().equals(other.disclaimers, _disclaimers)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,fiscalYear,section,const DeepCollectionEquality().hash(_lines),legend,const DeepCollectionEquality().hash(_disclaimers),engineVersion);
}

@override
String toString() {
    return 'AnnualBreakdown(taxpayerId: $taxpayerId, fiscalYear: $fiscalYear, section: $section, lines: $lines, legend: $legend, disclaimers: $disclaimers, engineVersion: $engineVersion)';
}


}

/// @nodoc
abstract mixin class _$AnnualBreakdownCopyWith<$Res> implements $AnnualBreakdownCopyWith<$Res> {
  factory _$AnnualBreakdownCopyWith(_AnnualBreakdown value, $Res Function(_AnnualBreakdown) _then) = __$AnnualBreakdownCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, AnnualSection section, List<AnnualBreakdownLine> lines, AnnualLegend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'engine_version') EngineVersion engineVersion
});


@override $AnnualSectionCopyWith<$Res> get section;

}
/// @nodoc
class __$AnnualBreakdownCopyWithImpl<$Res>
    implements _$AnnualBreakdownCopyWith<$Res> {
  __$AnnualBreakdownCopyWithImpl(this._self, this._then);

  final _AnnualBreakdown _self;
  final $Res Function(_AnnualBreakdown) _then;

/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? section = null,Object? lines = null,Object? legend = null,Object? disclaimers = null,Object? engineVersion = null,}) {
  return _then(_AnnualBreakdown(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,section: null == section ? _self.section : section // ignore: cast_nullable_to_non_nullable
as AnnualSection,lines: null == lines ? _self._lines : lines // ignore: cast_nullable_to_non_nullable
as List<AnnualBreakdownLine>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as AnnualLegend,disclaimers: null == disclaimers ? _self._disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,
  ));
}

/// Create a copy of AnnualBreakdown
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualSectionCopyWith<$Res> get section {
  
  return $AnnualSectionCopyWith<$Res>(_self.section, (value) {
    return _then(_self.copyWith(section: value));
  });
}
}

// dart format on
