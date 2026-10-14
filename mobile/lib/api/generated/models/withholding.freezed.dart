// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'withholding.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Withholding {

@JsonKey(name: 'isr_cents') NonNegativeCents get isrCents;@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;/// Fraction of `amount_cents` (RESICO `0.0125`, fees and rent `0.10`); `'0'` when no ISR is withheld.
@JsonKey(name: 'isr_rate') DecimalRate get isrRate;@JsonKey(name: 'iva_basis') WithholdingBasis get ivaBasis; Withholder get withholder;@JsonKey(name: 'net_cents') NonNegativeCents get netCents;/// Expenses with `withholder: you` and an amount above 0: the 17th of the month after `paid_on`, when the taxpayer must remit both withholdings (never netted against IVA). `null` otherwise.
@JsonKey(includeIfNull: true, name: 'remit_due_on') String? get remitDueOn;@JsonKey(name: 'rule_ids') List<WithholdingRuleId> get ruleIds;/// One line per entry of `rule_ids`, same order.
 List<WithholdingExplanation> get explanations;/// Version of the decision table and rates that produced the figures (e.g. `2026.1`). A CPA answer or a new year's rates bumps it, and `cmd/backfill-withholdings` re-derives older records (MODULE-withholdings §5.3).
@JsonKey(name: 'table_version') String get tableVersion;/// A ⚠️ rule a CPA has not confirmed matched (warning `withholding_cpa_unverified`).
@JsonKey(name: 'cpa_unverified') bool get cpaUnverified;
/// Create a copy of Withholding
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$WithholdingCopyWith<Withholding> get copyWith => _$WithholdingCopyWithImpl<Withholding>(this as Withholding, _$identity);

  /// Serializes this Withholding to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Withholding;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Withholding&&(identical(other.isrCents, _this.isrCents) || other.isrCents == _this.isrCents)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.isrRate, _this.isrRate) || other.isrRate == _this.isrRate)&&(identical(other.ivaBasis, _this.ivaBasis) || other.ivaBasis == _this.ivaBasis)&&(identical(other.withholder, _this.withholder) || other.withholder == _this.withholder)&&(identical(other.netCents, _this.netCents) || other.netCents == _this.netCents)&&(identical(other.remitDueOn, _this.remitDueOn) || other.remitDueOn == _this.remitDueOn)&&const DeepCollectionEquality().equals(other.ruleIds, _this.ruleIds)&&const DeepCollectionEquality().equals(other.explanations, _this.explanations)&&(identical(other.tableVersion, _this.tableVersion) || other.tableVersion == _this.tableVersion)&&(identical(other.cpaUnverified, _this.cpaUnverified) || other.cpaUnverified == _this.cpaUnverified));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Withholding;
  return Object.hash(runtimeType,_this.isrCents,_this.ivaCents,_this.isrRate,_this.ivaBasis,_this.withholder,_this.netCents,_this.remitDueOn,const DeepCollectionEquality().hash(_this.ruleIds),const DeepCollectionEquality().hash(_this.explanations),_this.tableVersion,_this.cpaUnverified);
}

@override
String toString() {
  final _this = this as Withholding;
  return 'Withholding(isrCents: ${_this.isrCents}, ivaCents: ${_this.ivaCents}, isrRate: ${_this.isrRate}, ivaBasis: ${_this.ivaBasis}, withholder: ${_this.withholder}, netCents: ${_this.netCents}, remitDueOn: ${_this.remitDueOn}, ruleIds: ${_this.ruleIds}, explanations: ${_this.explanations}, tableVersion: ${_this.tableVersion}, cpaUnverified: ${_this.cpaUnverified})';
}


}

/// @nodoc
abstract mixin class $WithholdingCopyWith<$Res>  {
  factory $WithholdingCopyWith(Withholding value, $Res Function(Withholding) _then) = _$WithholdingCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'isr_cents') NonNegativeCents isrCents,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'isr_rate') DecimalRate isrRate,@JsonKey(name: 'iva_basis') WithholdingBasis ivaBasis, Withholder withholder,@JsonKey(name: 'net_cents') NonNegativeCents netCents,@JsonKey(includeIfNull: true, name: 'remit_due_on') String? remitDueOn,@JsonKey(name: 'rule_ids') List<WithholdingRuleId> ruleIds, List<WithholdingExplanation> explanations,@JsonKey(name: 'table_version') String tableVersion,@JsonKey(name: 'cpa_unverified') bool cpaUnverified
});




}
/// @nodoc
class _$WithholdingCopyWithImpl<$Res>
    implements $WithholdingCopyWith<$Res> {
  _$WithholdingCopyWithImpl(this._self, this._then);

  final Withholding _self;
  final $Res Function(Withholding) _then;

/// Create a copy of Withholding
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? isrCents = null,Object? ivaCents = null,Object? isrRate = null,Object? ivaBasis = null,Object? withholder = null,Object? netCents = null,Object? remitDueOn = freezed,Object? ruleIds = null,Object? explanations = null,Object? tableVersion = null,Object? cpaUnverified = null,}) {
  return _then(Withholding(
isrCents: null == isrCents ? _self.isrCents : isrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrRate: null == isrRate ? _self.isrRate : isrRate // ignore: cast_nullable_to_non_nullable
as DecimalRate,ivaBasis: null == ivaBasis ? _self.ivaBasis : ivaBasis // ignore: cast_nullable_to_non_nullable
as WithholdingBasis,withholder: null == withholder ? _self.withholder : withholder // ignore: cast_nullable_to_non_nullable
as Withholder,netCents: null == netCents ? _self.netCents : netCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,remitDueOn: freezed == remitDueOn ? _self.remitDueOn : remitDueOn // ignore: cast_nullable_to_non_nullable
as String?,ruleIds: null == ruleIds ? _self.ruleIds : ruleIds // ignore: cast_nullable_to_non_nullable
as List<WithholdingRuleId>,explanations: null == explanations ? _self.explanations : explanations // ignore: cast_nullable_to_non_nullable
as List<WithholdingExplanation>,tableVersion: null == tableVersion ? _self.tableVersion : tableVersion // ignore: cast_nullable_to_non_nullable
as String,cpaUnverified: null == cpaUnverified ? _self.cpaUnverified : cpaUnverified // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [Withholding].
extension WithholdingPatterns on Withholding {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Withholding value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Withholding() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Withholding value)  $default,){
final _that = this;
switch (_that) {
case _Withholding():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Withholding value)?  $default,){
final _that = this;
switch (_that) {
case _Withholding() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_cents')  NonNegativeCents isrCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_rate')  DecimalRate isrRate, @JsonKey(name: 'iva_basis')  WithholdingBasis ivaBasis,  Withholder withholder, @JsonKey(name: 'net_cents')  NonNegativeCents netCents, @JsonKey(includeIfNull: true, name: 'remit_due_on')  String? remitDueOn, @JsonKey(name: 'rule_ids')  List<WithholdingRuleId> ruleIds,  List<WithholdingExplanation> explanations, @JsonKey(name: 'table_version')  String tableVersion, @JsonKey(name: 'cpa_unverified')  bool cpaUnverified)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Withholding() when $default != null:
return $default(_that.isrCents,_that.ivaCents,_that.isrRate,_that.ivaBasis,_that.withholder,_that.netCents,_that.remitDueOn,_that.ruleIds,_that.explanations,_that.tableVersion,_that.cpaUnverified);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_cents')  NonNegativeCents isrCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_rate')  DecimalRate isrRate, @JsonKey(name: 'iva_basis')  WithholdingBasis ivaBasis,  Withholder withholder, @JsonKey(name: 'net_cents')  NonNegativeCents netCents, @JsonKey(includeIfNull: true, name: 'remit_due_on')  String? remitDueOn, @JsonKey(name: 'rule_ids')  List<WithholdingRuleId> ruleIds,  List<WithholdingExplanation> explanations, @JsonKey(name: 'table_version')  String tableVersion, @JsonKey(name: 'cpa_unverified')  bool cpaUnverified)  $default,) {final _that = this;
switch (_that) {
case _Withholding():
return $default(_that.isrCents,_that.ivaCents,_that.isrRate,_that.ivaBasis,_that.withholder,_that.netCents,_that.remitDueOn,_that.ruleIds,_that.explanations,_that.tableVersion,_that.cpaUnverified);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'isr_cents')  NonNegativeCents isrCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_rate')  DecimalRate isrRate, @JsonKey(name: 'iva_basis')  WithholdingBasis ivaBasis,  Withholder withholder, @JsonKey(name: 'net_cents')  NonNegativeCents netCents, @JsonKey(includeIfNull: true, name: 'remit_due_on')  String? remitDueOn, @JsonKey(name: 'rule_ids')  List<WithholdingRuleId> ruleIds,  List<WithholdingExplanation> explanations, @JsonKey(name: 'table_version')  String tableVersion, @JsonKey(name: 'cpa_unverified')  bool cpaUnverified)?  $default,) {final _that = this;
switch (_that) {
case _Withholding() when $default != null:
return $default(_that.isrCents,_that.ivaCents,_that.isrRate,_that.ivaBasis,_that.withholder,_that.netCents,_that.remitDueOn,_that.ruleIds,_that.explanations,_that.tableVersion,_that.cpaUnverified);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Withholding implements Withholding {
  const _Withholding({@JsonKey(name: 'isr_cents') required this.isrCents, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(name: 'isr_rate') required this.isrRate, @JsonKey(name: 'iva_basis') required this.ivaBasis, required this.withholder, @JsonKey(name: 'net_cents') required this.netCents, @JsonKey(includeIfNull: true, name: 'remit_due_on') required this.remitDueOn, @JsonKey(name: 'rule_ids') required  List<WithholdingRuleId> ruleIds, required  List<WithholdingExplanation> explanations, @JsonKey(name: 'table_version') required this.tableVersion, @JsonKey(name: 'cpa_unverified') required this.cpaUnverified}): _ruleIds = ruleIds,_explanations = explanations;
  factory _Withholding.fromJson(Map<String, dynamic> json) => _$WithholdingFromJson(json);

@override@JsonKey(name: 'isr_cents') final  NonNegativeCents isrCents;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
/// Fraction of `amount_cents` (RESICO `0.0125`, fees and rent `0.10`); `'0'` when no ISR is withheld.
@override@JsonKey(name: 'isr_rate') final  DecimalRate isrRate;
@override@JsonKey(name: 'iva_basis') final  WithholdingBasis ivaBasis;
@override final  Withholder withholder;
@override@JsonKey(name: 'net_cents') final  NonNegativeCents netCents;
/// Expenses with `withholder: you` and an amount above 0: the 17th of the month after `paid_on`, when the taxpayer must remit both withholdings (never netted against IVA). `null` otherwise.
@override@JsonKey(includeIfNull: true, name: 'remit_due_on') final  String? remitDueOn;
 final  List<WithholdingRuleId> _ruleIds;
@override@JsonKey(name: 'rule_ids') List<WithholdingRuleId> get ruleIds {
  if (_ruleIds is EqualUnmodifiableListView) return _ruleIds;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_ruleIds);
}

/// One line per entry of `rule_ids`, same order.
 final  List<WithholdingExplanation> _explanations;
/// One line per entry of `rule_ids`, same order.
@override List<WithholdingExplanation> get explanations {
  if (_explanations is EqualUnmodifiableListView) return _explanations;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_explanations);
}

/// Version of the decision table and rates that produced the figures (e.g. `2026.1`). A CPA answer or a new year's rates bumps it, and `cmd/backfill-withholdings` re-derives older records (MODULE-withholdings §5.3).
@override@JsonKey(name: 'table_version') final  String tableVersion;
/// A ⚠️ rule a CPA has not confirmed matched (warning `withholding_cpa_unverified`).
@override@JsonKey(name: 'cpa_unverified') final  bool cpaUnverified;

/// Create a copy of Withholding
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$WithholdingCopyWith<_Withholding> get copyWith => __$WithholdingCopyWithImpl<_Withholding>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$WithholdingToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Withholding&&(identical(other.isrCents, isrCents) || other.isrCents == isrCents)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.isrRate, isrRate) || other.isrRate == isrRate)&&(identical(other.ivaBasis, ivaBasis) || other.ivaBasis == ivaBasis)&&(identical(other.withholder, withholder) || other.withholder == withholder)&&(identical(other.netCents, netCents) || other.netCents == netCents)&&(identical(other.remitDueOn, remitDueOn) || other.remitDueOn == remitDueOn)&&const DeepCollectionEquality().equals(other.ruleIds, _ruleIds)&&const DeepCollectionEquality().equals(other.explanations, _explanations)&&(identical(other.tableVersion, tableVersion) || other.tableVersion == tableVersion)&&(identical(other.cpaUnverified, cpaUnverified) || other.cpaUnverified == cpaUnverified));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,isrCents,ivaCents,isrRate,ivaBasis,withholder,netCents,remitDueOn,const DeepCollectionEquality().hash(_ruleIds),const DeepCollectionEquality().hash(_explanations),tableVersion,cpaUnverified);
}

@override
String toString() {
    return 'Withholding(isrCents: $isrCents, ivaCents: $ivaCents, isrRate: $isrRate, ivaBasis: $ivaBasis, withholder: $withholder, netCents: $netCents, remitDueOn: $remitDueOn, ruleIds: $ruleIds, explanations: $explanations, tableVersion: $tableVersion, cpaUnverified: $cpaUnverified)';
}


}

/// @nodoc
abstract mixin class _$WithholdingCopyWith<$Res> implements $WithholdingCopyWith<$Res> {
  factory _$WithholdingCopyWith(_Withholding value, $Res Function(_Withholding) _then) = __$WithholdingCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'isr_cents') NonNegativeCents isrCents,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'isr_rate') DecimalRate isrRate,@JsonKey(name: 'iva_basis') WithholdingBasis ivaBasis, Withholder withholder,@JsonKey(name: 'net_cents') NonNegativeCents netCents,@JsonKey(includeIfNull: true, name: 'remit_due_on') String? remitDueOn,@JsonKey(name: 'rule_ids') List<WithholdingRuleId> ruleIds, List<WithholdingExplanation> explanations,@JsonKey(name: 'table_version') String tableVersion,@JsonKey(name: 'cpa_unverified') bool cpaUnverified
});




}
/// @nodoc
class __$WithholdingCopyWithImpl<$Res>
    implements _$WithholdingCopyWith<$Res> {
  __$WithholdingCopyWithImpl(this._self, this._then);

  final _Withholding _self;
  final $Res Function(_Withholding) _then;

/// Create a copy of Withholding
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? isrCents = null,Object? ivaCents = null,Object? isrRate = null,Object? ivaBasis = null,Object? withholder = null,Object? netCents = null,Object? remitDueOn = freezed,Object? ruleIds = null,Object? explanations = null,Object? tableVersion = null,Object? cpaUnverified = null,}) {
  return _then(_Withholding(
isrCents: null == isrCents ? _self.isrCents : isrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrRate: null == isrRate ? _self.isrRate : isrRate // ignore: cast_nullable_to_non_nullable
as DecimalRate,ivaBasis: null == ivaBasis ? _self.ivaBasis : ivaBasis // ignore: cast_nullable_to_non_nullable
as WithholdingBasis,withholder: null == withholder ? _self.withholder : withholder // ignore: cast_nullable_to_non_nullable
as Withholder,netCents: null == netCents ? _self.netCents : netCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,remitDueOn: freezed == remitDueOn ? _self.remitDueOn : remitDueOn // ignore: cast_nullable_to_non_nullable
as String?,ruleIds: null == ruleIds ? _self._ruleIds : ruleIds // ignore: cast_nullable_to_non_nullable
as List<WithholdingRuleId>,explanations: null == explanations ? _self._explanations : explanations // ignore: cast_nullable_to_non_nullable
as List<WithholdingExplanation>,tableVersion: null == tableVersion ? _self.tableVersion : tableVersion // ignore: cast_nullable_to_non_nullable
as String,cpaUnverified: null == cpaUnverified ? _self.cpaUnverified : cpaUnverified // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
