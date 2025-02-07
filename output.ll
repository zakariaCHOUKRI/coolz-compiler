%Object = type { i8* }
%Int = type { i8*, i32 }
%String = type { i8*, i8*, i32 }
%Bool = type { i8*, i1 }
%Main = type { i8* }

define %Int* @Main_main(%Main* %self) {
0:
	%1 = call %Int* @Int_new()
	%2 = bitcast %Int* %1 to %Int*
	%3 = getelementptr %Int, %Int* %2, i32 0, i32 1
	store i32 42, i32* %3
	ret %Int* %1
}

declare %Int* @Int_new()

define i32 @main() {
0:
	%1 = call %Main* @Main_new()
	%2 = call %Object* @Main_main(%Main* %1)
	ret i32 0
}

declare %Main* @Main_new()

declare %Object* @Main_main()
