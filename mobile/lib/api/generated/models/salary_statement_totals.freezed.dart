// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'salary_statement_totals.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SalaryStatementTotals {

@JsonKey(name: 'accumulable_cents') NonNegativeCents get accumulableCents;@JsonKey(name: 'exempt_cents') NonNegativeCents get exemptCents;@JsonKey(name: 'isr_withheld_cents') NonNegativeCents get isrWithheldCents;
/// Create a copy of SalaryStatementTotals
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SalaryStatementTotalsCopyWith<SalaryStatementTotals> get copyWith => _$SalaryStatementTotalsCopyWithImpl<SalaryStatementTotals>(this as SalaryStatementTotals, _$identity);

  /// Serializes this SalaryStatementTotals to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SalaryStatementTotals;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SalaryStatementTotals&&(identical(other.accumulableCents, _this.accumulableCents) || other.accumulableCents == _this.accumulableCents)&&(identical(other.exemptCents, _this.exemptCents) || other.exemptCents == _this.exemptCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SalaryStatementTotals;
  return Object.hash(runtimeType,_this.accumulableCents,_this.exemptCents,_this.isrWithheldCents);
}

@override
String toString() {
  final _this = this as SalaryStatementTotals;
  return 'SalaryStatementTotals(accumulableCents: ${_this.accumulableCents}, exemptCents: ${_this.exemptCents}, isrWithheldCents: ${_this.isrWithheldCents})';
}


}

/// @nodoc
abstract mixin class $SalaryStatementTotalsCopyWith<$Res>  {
  factory $SalaryStatementTotalsCopyWith(SalaryStatementTotals value, $Res Function(SalaryStatementTotals) _then) = _$SalaryStatementTotalsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents
});




}
/// @nodoc
class _$SalaryStatementTotalsCopyWithImpl<$Res>
    implements $SalaryStatementTotalsCopyWith<$Res> {
  _$SalaryStatementTotalsCopyWithImpl(this._self, this._then);

  final SalaryStatementTotals _self;
  final $Res Function(SalaryStatementTotals) _then;

/// Create a copy of SalaryStatementTotals
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,}) {
  return _then(SalaryStatementTotals(
accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [SalaryStatementTotals].
extension SalaryStatementTotalsPatterns on SalaryStatementTotals {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SalaryStatementTotals value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SalaryStatementTotals() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SalaryStatementTotals value)  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementTotals():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SalaryStatementTotals value)?  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementTotals() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SalaryStatementTotals() when $default != null:
return $default(_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents)  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementTotals():
return $default(_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents)?  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementTotals() when $default != null:
return $default(_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SalaryStatementTotals implements SalaryStatementTotals {
  const _SalaryStatementTotals({@JsonKey(name: 'accumulable_cents') required this.accumulableCents, @JsonKey(name: 'exempt_cents') required this.exemptCents, @JsonKey(name: 'isr_withheld_cents') required this.isrWithheldCents});
  factory _SalaryStatementTotals.fromJson(Map<String, dynamic> json) => _$SalaryStatementTotalsFromJson(json);

@override@JsonKey(name: 'accumulable_cents') final  NonNegativeCents accumulableCents;
@override@JsonKey(name: 'exempt_cents') final  NonNegativeCents exemptCents;
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents isrWithheldCents;

/// Create a copy of SalaryStatementTotals
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SalaryStatementTotalsCopyWith<_SalaryStatementTotals> get copyWith => __$SalaryStatementTotalsCopyWithImpl<_SalaryStatementTotals>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SalaryStatementTotalsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SalaryStatementTotals&&(identical(other.accumulableCents, accumulableCents) || other.accumulableCents == accumulableCents)&&(identical(other.exemptCents, exemptCents) || other.exemptCents == exemptCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,accumulableCents,exemptCents,isrWithheldCents);
}

@override
String toString() {
    return 'SalaryStatementTotals(accumulableCents: $accumulableCents, exemptCents: $exemptCents, isrWithheldCents: $isrWithheldCents)';
}


}

/// @nodoc
abstract mixin class _$SalaryStatementTotalsCopyWith<$Res> implements $SalaryStatementTotalsCopyWith<$Res> {
  factory _$SalaryStatementTotalsCopyWith(_SalaryStatementTotals value, $Res Function(_SalaryStatementTotals) _then) = __$SalaryStatementTotalsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents
});




}
/// @nodoc
class __$SalaryStatementTotalsCopyWithImpl<$Res>
    implements _$SalaryStatementTotalsCopyWith<$Res> {
  __$SalaryStatementTotalsCopyWithImpl(this._self, this._then);

  final _SalaryStatementTotals _self;
  final $Res Function(_SalaryStatementTotals) _then;

/// Create a copy of SalaryStatementTotals
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,}) {
  return _then(_SalaryStatementTotals(
accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
