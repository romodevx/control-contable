// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'filed_payment_entry.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$FiledPaymentEntry {

 int get month;@JsonKey(name: 'amount_cents') NonNegativeCents get amountCents;
/// Create a copy of FiledPaymentEntry
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$FiledPaymentEntryCopyWith<FiledPaymentEntry> get copyWith => _$FiledPaymentEntryCopyWithImpl<FiledPaymentEntry>(this as FiledPaymentEntry, _$identity);

  /// Serializes this FiledPaymentEntry to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as FiledPaymentEntry;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is FiledPaymentEntry&&(identical(other.month, _this.month) || other.month == _this.month)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as FiledPaymentEntry;
  return Object.hash(runtimeType,_this.month,_this.amountCents);
}

@override
String toString() {
  final _this = this as FiledPaymentEntry;
  return 'FiledPaymentEntry(month: ${_this.month}, amountCents: ${_this.amountCents})';
}


}

/// @nodoc
abstract mixin class $FiledPaymentEntryCopyWith<$Res>  {
  factory $FiledPaymentEntryCopyWith(FiledPaymentEntry value, $Res Function(FiledPaymentEntry) _then) = _$FiledPaymentEntryCopyWithImpl;
@useResult
$Res call({
 int month,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents
});




}
/// @nodoc
class _$FiledPaymentEntryCopyWithImpl<$Res>
    implements $FiledPaymentEntryCopyWith<$Res> {
  _$FiledPaymentEntryCopyWithImpl(this._self, this._then);

  final FiledPaymentEntry _self;
  final $Res Function(FiledPaymentEntry) _then;

/// Create a copy of FiledPaymentEntry
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? month = null,Object? amountCents = null,}) {
  return _then(FiledPaymentEntry(
month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [FiledPaymentEntry].
extension FiledPaymentEntryPatterns on FiledPaymentEntry {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _FiledPaymentEntry value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _FiledPaymentEntry() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _FiledPaymentEntry value)  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentEntry():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _FiledPaymentEntry value)?  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentEntry() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int month, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _FiledPaymentEntry() when $default != null:
return $default(_that.month,_that.amountCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int month, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentEntry():
return $default(_that.month,_that.amountCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int month, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents)?  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentEntry() when $default != null:
return $default(_that.month,_that.amountCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _FiledPaymentEntry implements FiledPaymentEntry {
  const _FiledPaymentEntry({required this.month, @JsonKey(name: 'amount_cents') required this.amountCents});
  factory _FiledPaymentEntry.fromJson(Map<String, dynamic> json) => _$FiledPaymentEntryFromJson(json);

@override final  int month;
@override@JsonKey(name: 'amount_cents') final  NonNegativeCents amountCents;

/// Create a copy of FiledPaymentEntry
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$FiledPaymentEntryCopyWith<_FiledPaymentEntry> get copyWith => __$FiledPaymentEntryCopyWithImpl<_FiledPaymentEntry>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$FiledPaymentEntryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _FiledPaymentEntry&&(identical(other.month, month) || other.month == month)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,month,amountCents);
}

@override
String toString() {
    return 'FiledPaymentEntry(month: $month, amountCents: $amountCents)';
}


}

/// @nodoc
abstract mixin class _$FiledPaymentEntryCopyWith<$Res> implements $FiledPaymentEntryCopyWith<$Res> {
  factory _$FiledPaymentEntryCopyWith(_FiledPaymentEntry value, $Res Function(_FiledPaymentEntry) _then) = __$FiledPaymentEntryCopyWithImpl;
@override @useResult
$Res call({
 int month,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents
});




}
/// @nodoc
class __$FiledPaymentEntryCopyWithImpl<$Res>
    implements _$FiledPaymentEntryCopyWith<$Res> {
  __$FiledPaymentEntryCopyWithImpl(this._self, this._then);

  final _FiledPaymentEntry _self;
  final $Res Function(_FiledPaymentEntry) _then;

/// Create a copy of FiledPaymentEntry
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? month = null,Object? amountCents = null,}) {
  return _then(_FiledPaymentEntry(
month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
