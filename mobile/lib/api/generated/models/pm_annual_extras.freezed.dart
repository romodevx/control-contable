// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'pm_annual_extras.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PmAnnualExtras {

@JsonKey(name: 'ptu_base_cents') MoneyCents get ptuBaseCents;@JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents') NonNegativeCents? get ptuToDistributeCents;@JsonKey(includeIfNull: true, name: 'suggested_cu_next_year') DecimalRate? get suggestedCuNextYear;
/// Create a copy of PmAnnualExtras
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PmAnnualExtrasCopyWith<PmAnnualExtras> get copyWith => _$PmAnnualExtrasCopyWithImpl<PmAnnualExtras>(this as PmAnnualExtras, _$identity);

  /// Serializes this PmAnnualExtras to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PmAnnualExtras;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PmAnnualExtras&&(identical(other.ptuBaseCents, _this.ptuBaseCents) || other.ptuBaseCents == _this.ptuBaseCents)&&(identical(other.ptuToDistributeCents, _this.ptuToDistributeCents) || other.ptuToDistributeCents == _this.ptuToDistributeCents)&&(identical(other.suggestedCuNextYear, _this.suggestedCuNextYear) || other.suggestedCuNextYear == _this.suggestedCuNextYear));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PmAnnualExtras;
  return Object.hash(runtimeType,_this.ptuBaseCents,_this.ptuToDistributeCents,_this.suggestedCuNextYear);
}

@override
String toString() {
  final _this = this as PmAnnualExtras;
  return 'PmAnnualExtras(ptuBaseCents: ${_this.ptuBaseCents}, ptuToDistributeCents: ${_this.ptuToDistributeCents}, suggestedCuNextYear: ${_this.suggestedCuNextYear})';
}


}

/// @nodoc
abstract mixin class $PmAnnualExtrasCopyWith<$Res>  {
  factory $PmAnnualExtrasCopyWith(PmAnnualExtras value, $Res Function(PmAnnualExtras) _then) = _$PmAnnualExtrasCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'ptu_base_cents') MoneyCents ptuBaseCents,@JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents') NonNegativeCents? ptuToDistributeCents,@JsonKey(includeIfNull: true, name: 'suggested_cu_next_year') DecimalRate? suggestedCuNextYear
});




}
/// @nodoc
class _$PmAnnualExtrasCopyWithImpl<$Res>
    implements $PmAnnualExtrasCopyWith<$Res> {
  _$PmAnnualExtrasCopyWithImpl(this._self, this._then);

  final PmAnnualExtras _self;
  final $Res Function(PmAnnualExtras) _then;

/// Create a copy of PmAnnualExtras
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ptuBaseCents = null,Object? ptuToDistributeCents = freezed,Object? suggestedCuNextYear = freezed,}) {
  return _then(PmAnnualExtras(
ptuBaseCents: null == ptuBaseCents ? _self.ptuBaseCents : ptuBaseCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,ptuToDistributeCents: freezed == ptuToDistributeCents ? _self.ptuToDistributeCents : ptuToDistributeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,suggestedCuNextYear: freezed == suggestedCuNextYear ? _self.suggestedCuNextYear : suggestedCuNextYear // ignore: cast_nullable_to_non_nullable
as DecimalRate?,
  ));
}

}


/// Adds pattern-matching-related methods to [PmAnnualExtras].
extension PmAnnualExtrasPatterns on PmAnnualExtras {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PmAnnualExtras value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PmAnnualExtras() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PmAnnualExtras value)  $default,){
final _that = this;
switch (_that) {
case _PmAnnualExtras():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PmAnnualExtras value)?  $default,){
final _that = this;
switch (_that) {
case _PmAnnualExtras() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'ptu_base_cents')  MoneyCents ptuBaseCents, @JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents')  NonNegativeCents? ptuToDistributeCents, @JsonKey(includeIfNull: true, name: 'suggested_cu_next_year')  DecimalRate? suggestedCuNextYear)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PmAnnualExtras() when $default != null:
return $default(_that.ptuBaseCents,_that.ptuToDistributeCents,_that.suggestedCuNextYear);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'ptu_base_cents')  MoneyCents ptuBaseCents, @JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents')  NonNegativeCents? ptuToDistributeCents, @JsonKey(includeIfNull: true, name: 'suggested_cu_next_year')  DecimalRate? suggestedCuNextYear)  $default,) {final _that = this;
switch (_that) {
case _PmAnnualExtras():
return $default(_that.ptuBaseCents,_that.ptuToDistributeCents,_that.suggestedCuNextYear);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'ptu_base_cents')  MoneyCents ptuBaseCents, @JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents')  NonNegativeCents? ptuToDistributeCents, @JsonKey(includeIfNull: true, name: 'suggested_cu_next_year')  DecimalRate? suggestedCuNextYear)?  $default,) {final _that = this;
switch (_that) {
case _PmAnnualExtras() when $default != null:
return $default(_that.ptuBaseCents,_that.ptuToDistributeCents,_that.suggestedCuNextYear);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PmAnnualExtras implements PmAnnualExtras {
  const _PmAnnualExtras({@JsonKey(name: 'ptu_base_cents') required this.ptuBaseCents, @JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents') required this.ptuToDistributeCents, @JsonKey(includeIfNull: true, name: 'suggested_cu_next_year') required this.suggestedCuNextYear});
  factory _PmAnnualExtras.fromJson(Map<String, dynamic> json) => _$PmAnnualExtrasFromJson(json);

@override@JsonKey(name: 'ptu_base_cents') final  MoneyCents ptuBaseCents;
@override@JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents') final  NonNegativeCents? ptuToDistributeCents;
@override@JsonKey(includeIfNull: true, name: 'suggested_cu_next_year') final  DecimalRate? suggestedCuNextYear;

/// Create a copy of PmAnnualExtras
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PmAnnualExtrasCopyWith<_PmAnnualExtras> get copyWith => __$PmAnnualExtrasCopyWithImpl<_PmAnnualExtras>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PmAnnualExtrasToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PmAnnualExtras&&(identical(other.ptuBaseCents, ptuBaseCents) || other.ptuBaseCents == ptuBaseCents)&&(identical(other.ptuToDistributeCents, ptuToDistributeCents) || other.ptuToDistributeCents == ptuToDistributeCents)&&(identical(other.suggestedCuNextYear, suggestedCuNextYear) || other.suggestedCuNextYear == suggestedCuNextYear));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,ptuBaseCents,ptuToDistributeCents,suggestedCuNextYear);
}

@override
String toString() {
    return 'PmAnnualExtras(ptuBaseCents: $ptuBaseCents, ptuToDistributeCents: $ptuToDistributeCents, suggestedCuNextYear: $suggestedCuNextYear)';
}


}

/// @nodoc
abstract mixin class _$PmAnnualExtrasCopyWith<$Res> implements $PmAnnualExtrasCopyWith<$Res> {
  factory _$PmAnnualExtrasCopyWith(_PmAnnualExtras value, $Res Function(_PmAnnualExtras) _then) = __$PmAnnualExtrasCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'ptu_base_cents') MoneyCents ptuBaseCents,@JsonKey(includeIfNull: true, name: 'ptu_to_distribute_cents') NonNegativeCents? ptuToDistributeCents,@JsonKey(includeIfNull: true, name: 'suggested_cu_next_year') DecimalRate? suggestedCuNextYear
});




}
/// @nodoc
class __$PmAnnualExtrasCopyWithImpl<$Res>
    implements _$PmAnnualExtrasCopyWith<$Res> {
  __$PmAnnualExtrasCopyWithImpl(this._self, this._then);

  final _PmAnnualExtras _self;
  final $Res Function(_PmAnnualExtras) _then;

/// Create a copy of PmAnnualExtras
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ptuBaseCents = null,Object? ptuToDistributeCents = freezed,Object? suggestedCuNextYear = freezed,}) {
  return _then(_PmAnnualExtras(
ptuBaseCents: null == ptuBaseCents ? _self.ptuBaseCents : ptuBaseCents // ignore: cast_nullable_to_non_nullable
as MoneyCents,ptuToDistributeCents: freezed == ptuToDistributeCents ? _self.ptuToDistributeCents : ptuToDistributeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,suggestedCuNextYear: freezed == suggestedCuNextYear ? _self.suggestedCuNextYear : suggestedCuNextYear // ignore: cast_nullable_to_non_nullable
as DecimalRate?,
  ));
}


}

// dart format on
