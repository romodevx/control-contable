// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_page.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomePage {

 List<Income> get items;/// Opaque cursor of the next page; `null` on the last page.
@JsonKey(includeIfNull: true, name: 'next_cursor') String? get nextCursor; PeriodTotals get totals;
/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomePageCopyWith<IncomePage> get copyWith => _$IncomePageCopyWithImpl<IncomePage>(this as IncomePage, _$identity);

  /// Serializes this IncomePage to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomePage;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomePage&&const DeepCollectionEquality().equals(other.items, _this.items)&&(identical(other.nextCursor, _this.nextCursor) || other.nextCursor == _this.nextCursor)&&(identical(other.totals, _this.totals) || other.totals == _this.totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomePage;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items),_this.nextCursor,_this.totals);
}

@override
String toString() {
  final _this = this as IncomePage;
  return 'IncomePage(items: ${_this.items}, nextCursor: ${_this.nextCursor}, totals: ${_this.totals})';
}


}

/// @nodoc
abstract mixin class $IncomePageCopyWith<$Res>  {
  factory $IncomePageCopyWith(IncomePage value, $Res Function(IncomePage) _then) = _$IncomePageCopyWithImpl;
@useResult
$Res call({
 List<Income> items,@JsonKey(includeIfNull: true, name: 'next_cursor') String? nextCursor, PeriodTotals totals
});


$PeriodTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class _$IncomePageCopyWithImpl<$Res>
    implements $IncomePageCopyWith<$Res> {
  _$IncomePageCopyWithImpl(this._self, this._then);

  final IncomePage _self;
  final $Res Function(IncomePage) _then;

/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,Object? nextCursor = freezed,Object? totals = null,}) {
  return _then(IncomePage(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<Income>,nextCursor: freezed == nextCursor ? _self.nextCursor : nextCursor // ignore: cast_nullable_to_non_nullable
as String?,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as PeriodTotals,
  ));
}
/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PeriodTotalsCopyWith<$Res> get totals {
  
  return $PeriodTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}


/// Adds pattern-matching-related methods to [IncomePage].
extension IncomePagePatterns on IncomePage {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomePage value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomePage() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomePage value)  $default,){
final _that = this;
switch (_that) {
case _IncomePage():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomePage value)?  $default,){
final _that = this;
switch (_that) {
case _IncomePage() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Income> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomePage() when $default != null:
return $default(_that.items,_that.nextCursor,_that.totals);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Income> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals)  $default,) {final _that = this;
switch (_that) {
case _IncomePage():
return $default(_that.items,_that.nextCursor,_that.totals);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Income> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals)?  $default,) {final _that = this;
switch (_that) {
case _IncomePage() when $default != null:
return $default(_that.items,_that.nextCursor,_that.totals);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomePage implements IncomePage {
  const _IncomePage({required  List<Income> items, @JsonKey(includeIfNull: true, name: 'next_cursor') required this.nextCursor, required this.totals}): _items = items;
  factory _IncomePage.fromJson(Map<String, dynamic> json) => _$IncomePageFromJson(json);

 final  List<Income> _items;
@override List<Income> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}

/// Opaque cursor of the next page; `null` on the last page.
@override@JsonKey(includeIfNull: true, name: 'next_cursor') final  String? nextCursor;
@override final  PeriodTotals totals;

/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomePageCopyWith<_IncomePage> get copyWith => __$IncomePageCopyWithImpl<_IncomePage>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomePageToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomePage&&const DeepCollectionEquality().equals(other.items, _items)&&(identical(other.nextCursor, nextCursor) || other.nextCursor == nextCursor)&&(identical(other.totals, totals) || other.totals == totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items),nextCursor,totals);
}

@override
String toString() {
    return 'IncomePage(items: $items, nextCursor: $nextCursor, totals: $totals)';
}


}

/// @nodoc
abstract mixin class _$IncomePageCopyWith<$Res> implements $IncomePageCopyWith<$Res> {
  factory _$IncomePageCopyWith(_IncomePage value, $Res Function(_IncomePage) _then) = __$IncomePageCopyWithImpl;
@override @useResult
$Res call({
 List<Income> items,@JsonKey(includeIfNull: true, name: 'next_cursor') String? nextCursor, PeriodTotals totals
});


@override $PeriodTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class __$IncomePageCopyWithImpl<$Res>
    implements _$IncomePageCopyWith<$Res> {
  __$IncomePageCopyWithImpl(this._self, this._then);

  final _IncomePage _self;
  final $Res Function(_IncomePage) _then;

/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,Object? nextCursor = freezed,Object? totals = null,}) {
  return _then(_IncomePage(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<Income>,nextCursor: freezed == nextCursor ? _self.nextCursor : nextCursor // ignore: cast_nullable_to_non_nullable
as String?,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as PeriodTotals,
  ));
}

/// Create a copy of IncomePage
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PeriodTotalsCopyWith<$Res> get totals {
  
  return $PeriodTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}

// dart format on
