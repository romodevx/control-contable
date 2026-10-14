// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_last_used.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomeLastUsed {

@JsonKey(includeIfNull: true, name: 'counterparty_kind') CounterpartyKind? get counterpartyKind;@JsonKey(name: 'operation_kind') OperationKind get operationKind;@JsonKey(name: 'residential_rent') bool get residentialRent;@JsonKey(name: 'via_platform') bool get viaPlatform;
/// Create a copy of IncomeLastUsed
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeLastUsedCopyWith<IncomeLastUsed> get copyWith => _$IncomeLastUsedCopyWithImpl<IncomeLastUsed>(this as IncomeLastUsed, _$identity);

  /// Serializes this IncomeLastUsed to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomeLastUsed;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeLastUsed&&(identical(other.counterpartyKind, _this.counterpartyKind) || other.counterpartyKind == _this.counterpartyKind)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.viaPlatform, _this.viaPlatform) || other.viaPlatform == _this.viaPlatform));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeLastUsed;
  return Object.hash(runtimeType,_this.counterpartyKind,_this.operationKind,_this.residentialRent,_this.viaPlatform);
}

@override
String toString() {
  final _this = this as IncomeLastUsed;
  return 'IncomeLastUsed(counterpartyKind: ${_this.counterpartyKind}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, viaPlatform: ${_this.viaPlatform})';
}


}

/// @nodoc
abstract mixin class $IncomeLastUsedCopyWith<$Res>  {
  factory $IncomeLastUsedCopyWith(IncomeLastUsed value, $Res Function(IncomeLastUsed) _then) = _$IncomeLastUsedCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'counterparty_kind') CounterpartyKind? counterpartyKind,@JsonKey(name: 'operation_kind') OperationKind operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'via_platform') bool viaPlatform
});




}
/// @nodoc
class _$IncomeLastUsedCopyWithImpl<$Res>
    implements $IncomeLastUsedCopyWith<$Res> {
  _$IncomeLastUsedCopyWithImpl(this._self, this._then);

  final IncomeLastUsed _self;
  final $Res Function(IncomeLastUsed) _then;

/// Create a copy of IncomeLastUsed
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? counterpartyKind = freezed,Object? operationKind = null,Object? residentialRent = null,Object? viaPlatform = null,}) {
  return _then(IncomeLastUsed(
counterpartyKind: freezed == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind?,operationKind: null == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,viaPlatform: null == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [IncomeLastUsed].
extension IncomeLastUsedPatterns on IncomeLastUsed {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomeLastUsed value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomeLastUsed() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomeLastUsed value)  $default,){
final _that = this;
switch (_that) {
case _IncomeLastUsed():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomeLastUsed value)?  $default,){
final _that = this;
switch (_that) {
case _IncomeLastUsed() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeLastUsed() when $default != null:
return $default(_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform)  $default,) {final _that = this;
switch (_that) {
case _IncomeLastUsed():
return $default(_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: true, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform)?  $default,) {final _that = this;
switch (_that) {
case _IncomeLastUsed() when $default != null:
return $default(_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeLastUsed implements IncomeLastUsed {
  const _IncomeLastUsed({@JsonKey(includeIfNull: true, name: 'counterparty_kind') required this.counterpartyKind, @JsonKey(name: 'operation_kind') required this.operationKind, @JsonKey(name: 'residential_rent') required this.residentialRent, @JsonKey(name: 'via_platform') required this.viaPlatform});
  factory _IncomeLastUsed.fromJson(Map<String, dynamic> json) => _$IncomeLastUsedFromJson(json);

@override@JsonKey(includeIfNull: true, name: 'counterparty_kind') final  CounterpartyKind? counterpartyKind;
@override@JsonKey(name: 'operation_kind') final  OperationKind operationKind;
@override@JsonKey(name: 'residential_rent') final  bool residentialRent;
@override@JsonKey(name: 'via_platform') final  bool viaPlatform;

/// Create a copy of IncomeLastUsed
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeLastUsedCopyWith<_IncomeLastUsed> get copyWith => __$IncomeLastUsedCopyWithImpl<_IncomeLastUsed>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeLastUsedToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeLastUsed&&(identical(other.counterpartyKind, counterpartyKind) || other.counterpartyKind == counterpartyKind)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.viaPlatform, viaPlatform) || other.viaPlatform == viaPlatform));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,counterpartyKind,operationKind,residentialRent,viaPlatform);
}

@override
String toString() {
    return 'IncomeLastUsed(counterpartyKind: $counterpartyKind, operationKind: $operationKind, residentialRent: $residentialRent, viaPlatform: $viaPlatform)';
}


}

/// @nodoc
abstract mixin class _$IncomeLastUsedCopyWith<$Res> implements $IncomeLastUsedCopyWith<$Res> {
  factory _$IncomeLastUsedCopyWith(_IncomeLastUsed value, $Res Function(_IncomeLastUsed) _then) = __$IncomeLastUsedCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'counterparty_kind') CounterpartyKind? counterpartyKind,@JsonKey(name: 'operation_kind') OperationKind operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'via_platform') bool viaPlatform
});




}
/// @nodoc
class __$IncomeLastUsedCopyWithImpl<$Res>
    implements _$IncomeLastUsedCopyWith<$Res> {
  __$IncomeLastUsedCopyWithImpl(this._self, this._then);

  final _IncomeLastUsed _self;
  final $Res Function(_IncomeLastUsed) _then;

/// Create a copy of IncomeLastUsed
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? counterpartyKind = freezed,Object? operationKind = null,Object? residentialRent = null,Object? viaPlatform = null,}) {
  return _then(_IncomeLastUsed(
counterpartyKind: freezed == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind?,operationKind: null == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,viaPlatform: null == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
