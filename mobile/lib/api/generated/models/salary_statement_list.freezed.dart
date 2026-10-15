// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'salary_statement_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SalaryStatementList {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; List<SalaryStatement> get items; SalaryStatementTotals get totals;
/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SalaryStatementListCopyWith<SalaryStatementList> get copyWith => _$SalaryStatementListCopyWithImpl<SalaryStatementList>(this as SalaryStatementList, _$identity);

  /// Serializes this SalaryStatementList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SalaryStatementList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SalaryStatementList&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&const DeepCollectionEquality().equals(other.items, _this.items)&&(identical(other.totals, _this.totals) || other.totals == _this.totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SalaryStatementList;
  return Object.hash(runtimeType,_this.fiscalYear,const DeepCollectionEquality().hash(_this.items),_this.totals);
}

@override
String toString() {
  final _this = this as SalaryStatementList;
  return 'SalaryStatementList(fiscalYear: ${_this.fiscalYear}, items: ${_this.items}, totals: ${_this.totals})';
}


}

/// @nodoc
abstract mixin class $SalaryStatementListCopyWith<$Res>  {
  factory $SalaryStatementListCopyWith(SalaryStatementList value, $Res Function(SalaryStatementList) _then) = _$SalaryStatementListCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<SalaryStatement> items, SalaryStatementTotals totals
});


$SalaryStatementTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class _$SalaryStatementListCopyWithImpl<$Res>
    implements $SalaryStatementListCopyWith<$Res> {
  _$SalaryStatementListCopyWithImpl(this._self, this._then);

  final SalaryStatementList _self;
  final $Res Function(SalaryStatementList) _then;

/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? items = null,Object? totals = null,}) {
  return _then(SalaryStatementList(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<SalaryStatement>,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as SalaryStatementTotals,
  ));
}
/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SalaryStatementTotalsCopyWith<$Res> get totals {
  
  return $SalaryStatementTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}


/// Adds pattern-matching-related methods to [SalaryStatementList].
extension SalaryStatementListPatterns on SalaryStatementList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SalaryStatementList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SalaryStatementList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SalaryStatementList value)  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SalaryStatementList value)?  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<SalaryStatement> items,  SalaryStatementTotals totals)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SalaryStatementList() when $default != null:
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<SalaryStatement> items,  SalaryStatementTotals totals)  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementList():
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<SalaryStatement> items,  SalaryStatementTotals totals)?  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementList() when $default != null:
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SalaryStatementList implements SalaryStatementList {
  const _SalaryStatementList({@JsonKey(name: 'fiscal_year') required this.fiscalYear, required  List<SalaryStatement> items, required this.totals}): _items = items;
  factory _SalaryStatementList.fromJson(Map<String, dynamic> json) => _$SalaryStatementListFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
 final  List<SalaryStatement> _items;
@override List<SalaryStatement> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}

@override final  SalaryStatementTotals totals;

/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SalaryStatementListCopyWith<_SalaryStatementList> get copyWith => __$SalaryStatementListCopyWithImpl<_SalaryStatementList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SalaryStatementListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SalaryStatementList&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&const DeepCollectionEquality().equals(other.items, _items)&&(identical(other.totals, totals) || other.totals == totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,const DeepCollectionEquality().hash(_items),totals);
}

@override
String toString() {
    return 'SalaryStatementList(fiscalYear: $fiscalYear, items: $items, totals: $totals)';
}


}

/// @nodoc
abstract mixin class _$SalaryStatementListCopyWith<$Res> implements $SalaryStatementListCopyWith<$Res> {
  factory _$SalaryStatementListCopyWith(_SalaryStatementList value, $Res Function(_SalaryStatementList) _then) = __$SalaryStatementListCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<SalaryStatement> items, SalaryStatementTotals totals
});


@override $SalaryStatementTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class __$SalaryStatementListCopyWithImpl<$Res>
    implements _$SalaryStatementListCopyWith<$Res> {
  __$SalaryStatementListCopyWithImpl(this._self, this._then);

  final _SalaryStatementList _self;
  final $Res Function(_SalaryStatementList) _then;

/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? items = null,Object? totals = null,}) {
  return _then(_SalaryStatementList(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<SalaryStatement>,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as SalaryStatementTotals,
  ));
}

/// Create a copy of SalaryStatementList
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SalaryStatementTotalsCopyWith<$Res> get totals {
  
  return $SalaryStatementTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}

// dart format on
