%Object = type { i8* }
%Int = type { i8*, i32 }
%String = type { i8*, i8*, i32 }
%Bool = type { i8*, i1 }
%IO = type { i8* }
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
entry:
	%0 = call %Main* @Main_new()
	%1 = call %Int* @Main_main(%Main* %0)
	%2 = bitcast %Int* %1 to %Int*
	%3 = getelementptr %Int, %Int* %2, i32 0, i32 1
	%4 = load i32, i32* %3
	call void @print_int(i32 %4)
	ret i32 0
}

declare %Main* @Main_new()

declare void @print_int(i32 %n)
