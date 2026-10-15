// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_balance.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualBalance {

 AnnualBalanceKind get kind;@JsonKey(name: 'amount_cents') NonNegativeCents get amountCents;
/// Create a copy of AnnualBalance
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualBalanceCopyWith<AnnualBalance> get copyWith => _$AnnualBalanceCopyWithImpl<AnnualBalance>(this as AnnualBalance, _$identity);

  /// Serializes this AnnualBalance to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualBalance;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualBalance&&(identical(other.kind, _this.kind) || other.kind == _this.kind)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualBalance;
  return Object.hash(runtimeType,_this.kind,_this.amountCents);
}

@override
String toString() {
  final _this = this as AnnualBalance;
  return 'AnnualBalance(kind: ${_this.kind}, amountCents: ${_this.amountCents})';
}


}

/// @nodoc
abstract mixin class $AnnualBalanceCopyWith<$Res>  {
  factory $AnnualBalanceCopyWith(AnnualBalance value, $Res Function(AnnualBalance) _then) = _$AnnualBalanceCopyWithImpl;
@useResult
$Res call({
 AnnualBalanceKind kind,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents
});




}
/// @nodoc
class _$AnnualBalanceCopyWithImpl<$Res>
    implements $AnnualBalanceCopyWith<$Res> {
  _$AnnualBalanceCopyWithImpl(this._self, this._then);

  final AnnualBalance _self;
  final $Res Function(AnnualBalance) _then;

/// Create a copy of AnnualBalance
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? kind = null,Object? amountCents = null,}) {
  return _then(AnnualBalance(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as AnnualBalanceKind,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualBalance].
extension AnnualBalancePatterns on AnnualBalance {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualBalance value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualBalance() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualBalance value)  $default,){
final _that = this;
switch (_that) {
case _AnnualBalance():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualBalance value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualBalance() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( AnnualBalanceKind kind, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualBalance() when $default != null:
return $default(_that.kind,_that.amountCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( AnnualBalanceKind kind, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)  $default,) {final _that = this;
switch (_that) {
case _AnnualBalance():
return $default(_that.kind,_that.amountCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( AnnualBalanceKind kind, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)?  $default,) {final _that = this;
switch (_that) {
case _AnnualBalance() when $default != null:
return $default(_that.kind,_that.amountCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualBalance implements AnnualBalance {
  const _AnnualBalance({required this.kind, @JsonKey(name: 'amount_cents') required this.amountCents});
  factory _AnnualBalance.fromJson(Map<String, dynamic> json) => _$AnnualBalanceFromJson(json);

@override final  AnnualBalanceKind kind;
@override@JsonKey(name: 'amount_cents') final  NonNegativeCents amountCents;

/// Create a copy of AnnualBalance
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualBalanceCopyWith<_AnnualBalance> get copyWith => __$AnnualBalanceCopyWithImpl<_AnnualBalance>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualBalanceToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualBalance&&(identical(other.kind, kind) || other.kind == kind)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,kind,amountCents);
}

@override
String toString() {
    return 'AnnualBalance(kind: $kind, amountCents: $amountCents)';
}


}

/// @nodoc
abstract mixin class _$AnnualBalanceCopyWith<$Res> implements $AnnualBalanceCopyWith<$Res> {
  factory _$AnnualBalanceCopyWith(_AnnualBalance value, $Res Function(_AnnualBalance) _then) = __$AnnualBalanceCopyWithImpl;
@override @useResult
$Res call({
 AnnualBalanceKind kind,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents
});




}
/// @nodoc
class __$AnnualBalanceCopyWithImpl<$Res>
    implements _$AnnualBalanceCopyWith<$Res> {
  __$AnnualBalanceCopyWithImpl(this._self, this._then);

  final _AnnualBalance _self;
  final $Res Function(_AnnualBalance) _then;

/// Create a copy of AnnualBalance
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? kind = null,Object? amountCents = null,}) {
  return _then(_AnnualBalance(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as AnnualBalanceKind,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
