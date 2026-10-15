// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_inputs.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualInputs {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear;@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? get priorYearAccumulableIncomeCents;@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents get otherExemptIncomeCents;@JsonKey(name: 'told_employer_will_file') bool get toldEmployerWillFile;@JsonKey(name: 'regime_tax_inputs') List<AnnualRegimeTaxInput> get regimeTaxInputs;@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? get updatedAt;
/// Create a copy of AnnualInputs
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualInputsCopyWith<AnnualInputs> get copyWith => _$AnnualInputsCopyWithImpl<AnnualInputs>(this as AnnualInputs, _$identity);

  /// Serializes this AnnualInputs to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualInputs;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualInputs&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.priorYearAccumulableIncomeCents, _this.priorYearAccumulableIncomeCents) || other.priorYearAccumulableIncomeCents == _this.priorYearAccumulableIncomeCents)&&(identical(other.otherExemptIncomeCents, _this.otherExemptIncomeCents) || other.otherExemptIncomeCents == _this.otherExemptIncomeCents)&&(identical(other.toldEmployerWillFile, _this.toldEmployerWillFile) || other.toldEmployerWillFile == _this.toldEmployerWillFile)&&const DeepCollectionEquality().equals(other.regimeTaxInputs, _this.regimeTaxInputs)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualInputs;
  return Object.hash(runtimeType,_this.fiscalYear,_this.priorYearAccumulableIncomeCents,_this.otherExemptIncomeCents,_this.toldEmployerWillFile,const DeepCollectionEquality().hash(_this.regimeTaxInputs),_this.updatedAt);
}

@override
String toString() {
  final _this = this as AnnualInputs;
  return 'AnnualInputs(fiscalYear: ${_this.fiscalYear}, priorYearAccumulableIncomeCents: ${_this.priorYearAccumulableIncomeCents}, otherExemptIncomeCents: ${_this.otherExemptIncomeCents}, toldEmployerWillFile: ${_this.toldEmployerWillFile}, regimeTaxInputs: ${_this.regimeTaxInputs}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $AnnualInputsCopyWith<$Res>  {
  factory $AnnualInputsCopyWith(AnnualInputs value, $Res Function(AnnualInputs) _then) = _$AnnualInputsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? priorYearAccumulableIncomeCents,@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents otherExemptIncomeCents,@JsonKey(name: 'told_employer_will_file') bool toldEmployerWillFile,@JsonKey(name: 'regime_tax_inputs') List<AnnualRegimeTaxInput> regimeTaxInputs,@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? updatedAt
});




}
/// @nodoc
class _$AnnualInputsCopyWithImpl<$Res>
    implements $AnnualInputsCopyWith<$Res> {
  _$AnnualInputsCopyWithImpl(this._self, this._then);

  final AnnualInputs _self;
  final $Res Function(AnnualInputs) _then;

/// Create a copy of AnnualInputs
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? priorYearAccumulableIncomeCents = freezed,Object? otherExemptIncomeCents = null,Object? toldEmployerWillFile = null,Object? regimeTaxInputs = null,Object? updatedAt = freezed,}) {
  return _then(AnnualInputs(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,priorYearAccumulableIncomeCents: freezed == priorYearAccumulableIncomeCents ? _self.priorYearAccumulableIncomeCents : priorYearAccumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,otherExemptIncomeCents: null == otherExemptIncomeCents ? _self.otherExemptIncomeCents : otherExemptIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,toldEmployerWillFile: null == toldEmployerWillFile ? _self.toldEmployerWillFile : toldEmployerWillFile // ignore: cast_nullable_to_non_nullable
as bool,regimeTaxInputs: null == regimeTaxInputs ? _self.regimeTaxInputs : regimeTaxInputs // ignore: cast_nullable_to_non_nullable
as List<AnnualRegimeTaxInput>,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualInputs].
extension AnnualInputsPatterns on AnnualInputs {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualInputs value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualInputs() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualInputs value)  $default,){
final _that = this;
switch (_that) {
case _AnnualInputs():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualInputs value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualInputs() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile, @JsonKey(name: 'regime_tax_inputs')  List<AnnualRegimeTaxInput> regimeTaxInputs, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualInputs() when $default != null:
return $default(_that.fiscalYear,_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile,_that.regimeTaxInputs,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile, @JsonKey(name: 'regime_tax_inputs')  List<AnnualRegimeTaxInput> regimeTaxInputs, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)  $default,) {final _that = this;
switch (_that) {
case _AnnualInputs():
return $default(_that.fiscalYear,_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile,_that.regimeTaxInputs,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile, @JsonKey(name: 'regime_tax_inputs')  List<AnnualRegimeTaxInput> regimeTaxInputs, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _AnnualInputs() when $default != null:
return $default(_that.fiscalYear,_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile,_that.regimeTaxInputs,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualInputs implements AnnualInputs {
  const _AnnualInputs({@JsonKey(name: 'fiscal_year') required this.fiscalYear, @JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') required this.priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents') required this.otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file') required this.toldEmployerWillFile, @JsonKey(name: 'regime_tax_inputs') required  List<AnnualRegimeTaxInput> regimeTaxInputs, @JsonKey(includeIfNull: true, name: 'updated_at') required this.updatedAt}): _regimeTaxInputs = regimeTaxInputs;
  factory _AnnualInputs.fromJson(Map<String, dynamic> json) => _$AnnualInputsFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') final  NonNegativeCents? priorYearAccumulableIncomeCents;
@override@JsonKey(name: 'other_exempt_income_cents') final  NonNegativeCents otherExemptIncomeCents;
@override@JsonKey(name: 'told_employer_will_file') final  bool toldEmployerWillFile;
 final  List<AnnualRegimeTaxInput> _regimeTaxInputs;
@override@JsonKey(name: 'regime_tax_inputs') List<AnnualRegimeTaxInput> get regimeTaxInputs {
  if (_regimeTaxInputs is EqualUnmodifiableListView) return _regimeTaxInputs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimeTaxInputs);
}

@override@JsonKey(includeIfNull: true, name: 'updated_at') final  DateTime? updatedAt;

/// Create a copy of AnnualInputs
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualInputsCopyWith<_AnnualInputs> get copyWith => __$AnnualInputsCopyWithImpl<_AnnualInputs>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualInputsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualInputs&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.priorYearAccumulableIncomeCents, priorYearAccumulableIncomeCents) || other.priorYearAccumulableIncomeCents == priorYearAccumulableIncomeCents)&&(identical(other.otherExemptIncomeCents, otherExemptIncomeCents) || other.otherExemptIncomeCents == otherExemptIncomeCents)&&(identical(other.toldEmployerWillFile, toldEmployerWillFile) || other.toldEmployerWillFile == toldEmployerWillFile)&&const DeepCollectionEquality().equals(other.regimeTaxInputs, _regimeTaxInputs)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,priorYearAccumulableIncomeCents,otherExemptIncomeCents,toldEmployerWillFile,const DeepCollectionEquality().hash(_regimeTaxInputs),updatedAt);
}

@override
String toString() {
    return 'AnnualInputs(fiscalYear: $fiscalYear, priorYearAccumulableIncomeCents: $priorYearAccumulableIncomeCents, otherExemptIncomeCents: $otherExemptIncomeCents, toldEmployerWillFile: $toldEmployerWillFile, regimeTaxInputs: $regimeTaxInputs, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$AnnualInputsCopyWith<$Res> implements $AnnualInputsCopyWith<$Res> {
  factory _$AnnualInputsCopyWith(_AnnualInputs value, $Res Function(_AnnualInputs) _then) = __$AnnualInputsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? priorYearAccumulableIncomeCents,@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents otherExemptIncomeCents,@JsonKey(name: 'told_employer_will_file') bool toldEmployerWillFile,@JsonKey(name: 'regime_tax_inputs') List<AnnualRegimeTaxInput> regimeTaxInputs,@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? updatedAt
});




}
/// @nodoc
class __$AnnualInputsCopyWithImpl<$Res>
    implements _$AnnualInputsCopyWith<$Res> {
  __$AnnualInputsCopyWithImpl(this._self, this._then);

  final _AnnualInputs _self;
  final $Res Function(_AnnualInputs) _then;

/// Create a copy of AnnualInputs
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? priorYearAccumulableIncomeCents = freezed,Object? otherExemptIncomeCents = null,Object? toldEmployerWillFile = null,Object? regimeTaxInputs = null,Object? updatedAt = freezed,}) {
  return _then(_AnnualInputs(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,priorYearAccumulableIncomeCents: freezed == priorYearAccumulableIncomeCents ? _self.priorYearAccumulableIncomeCents : priorYearAccumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,otherExemptIncomeCents: null == otherExemptIncomeCents ? _self.otherExemptIncomeCents : otherExemptIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,toldEmployerWillFile: null == toldEmployerWillFile ? _self.toldEmployerWillFile : toldEmployerWillFile // ignore: cast_nullable_to_non_nullable
as bool,regimeTaxInputs: null == regimeTaxInputs ? _self._regimeTaxInputs : regimeTaxInputs // ignore: cast_nullable_to_non_nullable
as List<AnnualRegimeTaxInput>,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
