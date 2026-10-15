// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_calendar_entry.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxCalendarEntry {

 Period get period; TaxObligation get obligation;/// Day 17 of the month after `period`.
@JsonKey(name: 'nominal_due_date') String get nominalDueDate;/// The nominal date moved to the next business day when needed.
@JsonKey(name: 'due_date') String get dueDate;/// `null` when `due_date` equals `nominal_due_date`.
@JsonKey(includeIfNull: true, name: 'moved_reason') TaxCalendarMovedReason? get movedReason;/// `true` while the date rule awaits CPA confirmation (OQ-R1, the CFF 12 ¶5 Friday extension is not applied): today every entry. The app shows "sujeto a confirmación".
@JsonKey(name: 'subject_to_confirmation') bool get subjectToConfirmation;
/// Create a copy of TaxCalendarEntry
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxCalendarEntryCopyWith<TaxCalendarEntry> get copyWith => _$TaxCalendarEntryCopyWithImpl<TaxCalendarEntry>(this as TaxCalendarEntry, _$identity);

  /// Serializes this TaxCalendarEntry to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxCalendarEntry;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxCalendarEntry&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.obligation, _this.obligation) || other.obligation == _this.obligation)&&(identical(other.nominalDueDate, _this.nominalDueDate) || other.nominalDueDate == _this.nominalDueDate)&&(identical(other.dueDate, _this.dueDate) || other.dueDate == _this.dueDate)&&(identical(other.movedReason, _this.movedReason) || other.movedReason == _this.movedReason)&&(identical(other.subjectToConfirmation, _this.subjectToConfirmation) || other.subjectToConfirmation == _this.subjectToConfirmation));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxCalendarEntry;
  return Object.hash(runtimeType,_this.period,_this.obligation,_this.nominalDueDate,_this.dueDate,_this.movedReason,_this.subjectToConfirmation);
}

@override
String toString() {
  final _this = this as TaxCalendarEntry;
  return 'TaxCalendarEntry(period: ${_this.period}, obligation: ${_this.obligation}, nominalDueDate: ${_this.nominalDueDate}, dueDate: ${_this.dueDate}, movedReason: ${_this.movedReason}, subjectToConfirmation: ${_this.subjectToConfirmation})';
}


}

/// @nodoc
abstract mixin class $TaxCalendarEntryCopyWith<$Res>  {
  factory $TaxCalendarEntryCopyWith(TaxCalendarEntry value, $Res Function(TaxCalendarEntry) _then) = _$TaxCalendarEntryCopyWithImpl;
@useResult
$Res call({
 Period period, TaxObligation obligation,@JsonKey(name: 'nominal_due_date') String nominalDueDate,@JsonKey(name: 'due_date') String dueDate,@JsonKey(includeIfNull: true, name: 'moved_reason') TaxCalendarMovedReason? movedReason,@JsonKey(name: 'subject_to_confirmation') bool subjectToConfirmation
});




}
/// @nodoc
class _$TaxCalendarEntryCopyWithImpl<$Res>
    implements $TaxCalendarEntryCopyWith<$Res> {
  _$TaxCalendarEntryCopyWithImpl(this._self, this._then);

  final TaxCalendarEntry _self;
  final $Res Function(TaxCalendarEntry) _then;

/// Create a copy of TaxCalendarEntry
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? period = null,Object? obligation = null,Object? nominalDueDate = null,Object? dueDate = null,Object? movedReason = freezed,Object? subjectToConfirmation = null,}) {
  return _then(TaxCalendarEntry(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,obligation: null == obligation ? _self.obligation : obligation // ignore: cast_nullable_to_non_nullable
as TaxObligation,nominalDueDate: null == nominalDueDate ? _self.nominalDueDate : nominalDueDate // ignore: cast_nullable_to_non_nullable
as String,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,movedReason: freezed == movedReason ? _self.movedReason : movedReason // ignore: cast_nullable_to_non_nullable
as TaxCalendarMovedReason?,subjectToConfirmation: null == subjectToConfirmation ? _self.subjectToConfirmation : subjectToConfirmation // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxCalendarEntry].
extension TaxCalendarEntryPatterns on TaxCalendarEntry {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxCalendarEntry value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxCalendarEntry() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxCalendarEntry value)  $default,){
final _that = this;
switch (_that) {
case _TaxCalendarEntry():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxCalendarEntry value)?  $default,){
final _that = this;
switch (_that) {
case _TaxCalendarEntry() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Period period,  TaxObligation obligation, @JsonKey(name: 'nominal_due_date')  String nominalDueDate, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(includeIfNull: true, name: 'moved_reason')  TaxCalendarMovedReason? movedReason, @JsonKey(name: 'subject_to_confirmation')  bool subjectToConfirmation)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxCalendarEntry() when $default != null:
return $default(_that.period,_that.obligation,_that.nominalDueDate,_that.dueDate,_that.movedReason,_that.subjectToConfirmation);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Period period,  TaxObligation obligation, @JsonKey(name: 'nominal_due_date')  String nominalDueDate, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(includeIfNull: true, name: 'moved_reason')  TaxCalendarMovedReason? movedReason, @JsonKey(name: 'subject_to_confirmation')  bool subjectToConfirmation)  $default,) {final _that = this;
switch (_that) {
case _TaxCalendarEntry():
return $default(_that.period,_that.obligation,_that.nominalDueDate,_that.dueDate,_that.movedReason,_that.subjectToConfirmation);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Period period,  TaxObligation obligation, @JsonKey(name: 'nominal_due_date')  String nominalDueDate, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(includeIfNull: true, name: 'moved_reason')  TaxCalendarMovedReason? movedReason, @JsonKey(name: 'subject_to_confirmation')  bool subjectToConfirmation)?  $default,) {final _that = this;
switch (_that) {
case _TaxCalendarEntry() when $default != null:
return $default(_that.period,_that.obligation,_that.nominalDueDate,_that.dueDate,_that.movedReason,_that.subjectToConfirmation);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxCalendarEntry implements TaxCalendarEntry {
  const _TaxCalendarEntry({required this.period, required this.obligation, @JsonKey(name: 'nominal_due_date') required this.nominalDueDate, @JsonKey(name: 'due_date') required this.dueDate, @JsonKey(includeIfNull: true, name: 'moved_reason') required this.movedReason, @JsonKey(name: 'subject_to_confirmation') required this.subjectToConfirmation});
  factory _TaxCalendarEntry.fromJson(Map<String, dynamic> json) => _$TaxCalendarEntryFromJson(json);

@override final  Period period;
@override final  TaxObligation obligation;
/// Day 17 of the month after `period`.
@override@JsonKey(name: 'nominal_due_date') final  String nominalDueDate;
/// The nominal date moved to the next business day when needed.
@override@JsonKey(name: 'due_date') final  String dueDate;
/// `null` when `due_date` equals `nominal_due_date`.
@override@JsonKey(includeIfNull: true, name: 'moved_reason') final  TaxCalendarMovedReason? movedReason;
/// `true` while the date rule awaits CPA confirmation (OQ-R1, the CFF 12 ¶5 Friday extension is not applied): today every entry. The app shows "sujeto a confirmación".
@override@JsonKey(name: 'subject_to_confirmation') final  bool subjectToConfirmation;

/// Create a copy of TaxCalendarEntry
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxCalendarEntryCopyWith<_TaxCalendarEntry> get copyWith => __$TaxCalendarEntryCopyWithImpl<_TaxCalendarEntry>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxCalendarEntryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxCalendarEntry&&(identical(other.period, period) || other.period == period)&&(identical(other.obligation, obligation) || other.obligation == obligation)&&(identical(other.nominalDueDate, nominalDueDate) || other.nominalDueDate == nominalDueDate)&&(identical(other.dueDate, dueDate) || other.dueDate == dueDate)&&(identical(other.movedReason, movedReason) || other.movedReason == movedReason)&&(identical(other.subjectToConfirmation, subjectToConfirmation) || other.subjectToConfirmation == subjectToConfirmation));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,period,obligation,nominalDueDate,dueDate,movedReason,subjectToConfirmation);
}

@override
String toString() {
    return 'TaxCalendarEntry(period: $period, obligation: $obligation, nominalDueDate: $nominalDueDate, dueDate: $dueDate, movedReason: $movedReason, subjectToConfirmation: $subjectToConfirmation)';
}


}

/// @nodoc
abstract mixin class _$TaxCalendarEntryCopyWith<$Res> implements $TaxCalendarEntryCopyWith<$Res> {
  factory _$TaxCalendarEntryCopyWith(_TaxCalendarEntry value, $Res Function(_TaxCalendarEntry) _then) = __$TaxCalendarEntryCopyWithImpl;
@override @useResult
$Res call({
 Period period, TaxObligation obligation,@JsonKey(name: 'nominal_due_date') String nominalDueDate,@JsonKey(name: 'due_date') String dueDate,@JsonKey(includeIfNull: true, name: 'moved_reason') TaxCalendarMovedReason? movedReason,@JsonKey(name: 'subject_to_confirmation') bool subjectToConfirmation
});




}
/// @nodoc
class __$TaxCalendarEntryCopyWithImpl<$Res>
    implements _$TaxCalendarEntryCopyWith<$Res> {
  __$TaxCalendarEntryCopyWithImpl(this._self, this._then);

  final _TaxCalendarEntry _self;
  final $Res Function(_TaxCalendarEntry) _then;

/// Create a copy of TaxCalendarEntry
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? period = null,Object? obligation = null,Object? nominalDueDate = null,Object? dueDate = null,Object? movedReason = freezed,Object? subjectToConfirmation = null,}) {
  return _then(_TaxCalendarEntry(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,obligation: null == obligation ? _self.obligation : obligation // ignore: cast_nullable_to_non_nullable
as TaxObligation,nominalDueDate: null == nominalDueDate ? _self.nominalDueDate : nominalDueDate // ignore: cast_nullable_to_non_nullable
as String,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,movedReason: freezed == movedReason ? _self.movedReason : movedReason // ignore: cast_nullable_to_non_nullable
as TaxCalendarMovedReason?,subjectToConfirmation: null == subjectToConfirmation ? _self.subjectToConfirmation : subjectToConfirmation // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
