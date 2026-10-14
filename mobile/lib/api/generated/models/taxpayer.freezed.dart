// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Taxpayer {

 Uuid get id;@JsonKey(name: 'person_type') PersonType get personType;@JsonKey(name: 'display_name') String get displayName;/// Name part, six asterisks for the date, then homoclave and check digit (física `AAAA******XX9`, moral `AAA******XX9`). The full RFC is never returned.
@JsonKey(name: 'rfc_masked') String get rfcMasked;/// Every regime the taxpayer holds or held (ended ones carry `end_period`), ordered by `start_period`, then code.
 List<TaxpayerRegime> get regimes;@JsonKey(includeIfNull: true, name: 'archived_at') DateTime? get archivedAt;@JsonKey(name: 'created_at') DateTime get createdAt;/// Non-blocking notices about the input just written (`rfc_check_digit_mismatch`, `regime_combination_needs_cpa_review`). Present only on `createTaxpayer` (201 and 200) and `updateTaxpayer` responses, and only when non-empty; never stored, never on reads.
@JsonKey(includeIfNull: false) List<Warning>? get warnings;
/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerCopyWith<Taxpayer> get copyWith => _$TaxpayerCopyWithImpl<Taxpayer>(this as Taxpayer, _$identity);

  /// Serializes this Taxpayer to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Taxpayer;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Taxpayer&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.personType, _this.personType) || other.personType == _this.personType)&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName)&&(identical(other.rfcMasked, _this.rfcMasked) || other.rfcMasked == _this.rfcMasked)&&const DeepCollectionEquality().equals(other.regimes, _this.regimes)&&(identical(other.archivedAt, _this.archivedAt) || other.archivedAt == _this.archivedAt)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Taxpayer;
  return Object.hash(runtimeType,_this.id,_this.personType,_this.displayName,_this.rfcMasked,const DeepCollectionEquality().hash(_this.regimes),_this.archivedAt,_this.createdAt,const DeepCollectionEquality().hash(_this.warnings));
}

@override
String toString() {
  final _this = this as Taxpayer;
  return 'Taxpayer(id: ${_this.id}, personType: ${_this.personType}, displayName: ${_this.displayName}, rfcMasked: ${_this.rfcMasked}, regimes: ${_this.regimes}, archivedAt: ${_this.archivedAt}, createdAt: ${_this.createdAt}, warnings: ${_this.warnings})';
}


}

/// @nodoc
abstract mixin class $TaxpayerCopyWith<$Res>  {
  factory $TaxpayerCopyWith(Taxpayer value, $Res Function(Taxpayer) _then) = _$TaxpayerCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'person_type') PersonType personType,@JsonKey(name: 'display_name') String displayName,@JsonKey(name: 'rfc_masked') String rfcMasked, List<TaxpayerRegime> regimes,@JsonKey(includeIfNull: true, name: 'archived_at') DateTime? archivedAt,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(includeIfNull: false) List<Warning>? warnings
});




}
/// @nodoc
class _$TaxpayerCopyWithImpl<$Res>
    implements $TaxpayerCopyWith<$Res> {
  _$TaxpayerCopyWithImpl(this._self, this._then);

  final Taxpayer _self;
  final $Res Function(Taxpayer) _then;

/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? personType = null,Object? displayName = null,Object? rfcMasked = null,Object? regimes = null,Object? archivedAt = freezed,Object? createdAt = null,Object? warnings = freezed,}) {
  return _then(Taxpayer(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,rfcMasked: null == rfcMasked ? _self.rfcMasked : rfcMasked // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self.regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegime>,archivedAt: freezed == archivedAt ? _self.archivedAt : archivedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,warnings: freezed == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>?,
  ));
}

}


/// Adds pattern-matching-related methods to [Taxpayer].
extension TaxpayerPatterns on Taxpayer {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Taxpayer value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Taxpayer value)  $default,){
final _that = this;
switch (_that) {
case _Taxpayer():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Taxpayer value)?  $default,){
final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(includeIfNull: true, name: 'archived_at')  DateTime? archivedAt, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(includeIfNull: false)  List<Warning>? warnings)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.archivedAt,_that.createdAt,_that.warnings);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(includeIfNull: true, name: 'archived_at')  DateTime? archivedAt, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(includeIfNull: false)  List<Warning>? warnings)  $default,) {final _that = this;
switch (_that) {
case _Taxpayer():
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.archivedAt,_that.createdAt,_that.warnings);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(includeIfNull: true, name: 'archived_at')  DateTime? archivedAt, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(includeIfNull: false)  List<Warning>? warnings)?  $default,) {final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.archivedAt,_that.createdAt,_that.warnings);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Taxpayer implements Taxpayer {
  const _Taxpayer({required this.id, @JsonKey(name: 'person_type') required this.personType, @JsonKey(name: 'display_name') required this.displayName, @JsonKey(name: 'rfc_masked') required this.rfcMasked, required  List<TaxpayerRegime> regimes, @JsonKey(includeIfNull: true, name: 'archived_at') required this.archivedAt, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(includeIfNull: false)  List<Warning>? warnings}): _regimes = regimes,_warnings = warnings;
  factory _Taxpayer.fromJson(Map<String, dynamic> json) => _$TaxpayerFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'person_type') final  PersonType personType;
@override@JsonKey(name: 'display_name') final  String displayName;
/// Name part, six asterisks for the date, then homoclave and check digit (física `AAAA******XX9`, moral `AAA******XX9`). The full RFC is never returned.
@override@JsonKey(name: 'rfc_masked') final  String rfcMasked;
/// Every regime the taxpayer holds or held (ended ones carry `end_period`), ordered by `start_period`, then code.
 final  List<TaxpayerRegime> _regimes;
/// Every regime the taxpayer holds or held (ended ones carry `end_period`), ordered by `start_period`, then code.
@override List<TaxpayerRegime> get regimes {
  if (_regimes is EqualUnmodifiableListView) return _regimes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimes);
}

@override@JsonKey(includeIfNull: true, name: 'archived_at') final  DateTime? archivedAt;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
/// Non-blocking notices about the input just written (`rfc_check_digit_mismatch`, `regime_combination_needs_cpa_review`). Present only on `createTaxpayer` (201 and 200) and `updateTaxpayer` responses, and only when non-empty; never stored, never on reads.
 final  List<Warning>? _warnings;
/// Non-blocking notices about the input just written (`rfc_check_digit_mismatch`, `regime_combination_needs_cpa_review`). Present only on `createTaxpayer` (201 and 200) and `updateTaxpayer` responses, and only when non-empty; never stored, never on reads.
@override@JsonKey(includeIfNull: false) List<Warning>? get warnings {
  final value = _warnings;
  if (value == null) return null;
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}


/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerCopyWith<_Taxpayer> get copyWith => __$TaxpayerCopyWithImpl<_Taxpayer>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Taxpayer&&(identical(other.id, id) || other.id == id)&&(identical(other.personType, personType) || other.personType == personType)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&(identical(other.rfcMasked, rfcMasked) || other.rfcMasked == rfcMasked)&&const DeepCollectionEquality().equals(other.regimes, _regimes)&&(identical(other.archivedAt, archivedAt) || other.archivedAt == archivedAt)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&const DeepCollectionEquality().equals(other.warnings, _warnings));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,personType,displayName,rfcMasked,const DeepCollectionEquality().hash(_regimes),archivedAt,createdAt,const DeepCollectionEquality().hash(_warnings));
}

@override
String toString() {
    return 'Taxpayer(id: $id, personType: $personType, displayName: $displayName, rfcMasked: $rfcMasked, regimes: $regimes, archivedAt: $archivedAt, createdAt: $createdAt, warnings: $warnings)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerCopyWith<$Res> implements $TaxpayerCopyWith<$Res> {
  factory _$TaxpayerCopyWith(_Taxpayer value, $Res Function(_Taxpayer) _then) = __$TaxpayerCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'person_type') PersonType personType,@JsonKey(name: 'display_name') String displayName,@JsonKey(name: 'rfc_masked') String rfcMasked, List<TaxpayerRegime> regimes,@JsonKey(includeIfNull: true, name: 'archived_at') DateTime? archivedAt,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(includeIfNull: false) List<Warning>? warnings
});




}
/// @nodoc
class __$TaxpayerCopyWithImpl<$Res>
    implements _$TaxpayerCopyWith<$Res> {
  __$TaxpayerCopyWithImpl(this._self, this._then);

  final _Taxpayer _self;
  final $Res Function(_Taxpayer) _then;

/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? personType = null,Object? displayName = null,Object? rfcMasked = null,Object? regimes = null,Object? archivedAt = freezed,Object? createdAt = null,Object? warnings = freezed,}) {
  return _then(_Taxpayer(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,rfcMasked: null == rfcMasked ? _self.rfcMasked : rfcMasked // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self._regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegime>,archivedAt: freezed == archivedAt ? _self.archivedAt : archivedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,warnings: freezed == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>?,
  ));
}


}

// dart format on
