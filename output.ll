target triple = "x86_64-pc-windows-msvc19.43.34808"

@str.0 = global [55 x i8] c"Error: the program was aborted by an abort() function\0A\00"
@str.1 = global [7 x i8] c"Object\00"
@str.2 = global [3 x i8] c"%s\00"
@str.3 = global [5 x i8] c"%lld\00"
@str.4 = global [9 x i8] c"%255[^\0A]\00"
@str.5 = global [4 x i8] c"%*c\00"
@str.6 = global [8 x i8] c"Printer\00"
@str.7 = global [2 x i8] c"\0A\00"
@str.8 = global [5 x i8] c"Main\00"
@str.9 = global [15 x i8] c"Hello everyone\00"
@str.10 = global [15 x i8] c"the length of:\00"
@str.11 = global [2 x i8] c"\22\00"
@str.12 = global [3 x i8] c"\22 \00"
@str.13 = global [5 x i8] c"is: \00"

declare i32 @printf(i8* %format, ...)

declare i32 @scanf(i8* %format, ...)

declare i8* @memset(i8* %str, i32 %c, i64 %n)

declare i8* @malloc(i64 %size)

declare i8* @memcpy(i8* %dest, i8* %src, i64 %size)

declare i64 @strlen(i8* %str)

define i8* @Object_abort(i8* %self) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([55 x i8], [55 x i8]* @str.0, i32 0, i32 0))
	call void @exit(i32 1)
	unreachable
}

declare void @exit(i32 %status)

define i8* @Object_type_name(i8* %self) {
0:
	ret i8* getelementptr ([7 x i8], [7 x i8]* @str.1, i32 0, i32 0)
}

define i8* @Object_copy(i8* %self) {
0:
	%1 = call i8* @malloc(i64 8)
	%2 = call i8* @memcpy(i8* %1, i8* %self, i64 8)
	ret i8* %1
}

define i8* @IO_out_string(i8* %self, i8* %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([3 x i8], [3 x i8]* @str.2, i32 0, i32 0), i8* %x)
	ret i8* %self
}

define i8* @IO_out_int(i8* %self, i64 %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([5 x i8], [5 x i8]* @str.3, i32 0, i32 0), i64 %x)
	ret i8* %self
}

define i8* @IO_in_string(i8* %self) {
0:
	%1 = alloca [256 x i8]
	%2 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%3 = call i8* @memset(i8* %2, i32 0, i64 256)
	%4 = call i32 (i8*, ...) @scanf(i8* getelementptr ([9 x i8], [9 x i8]* @str.4, i32 0, i32 0), [256 x i8]* %1)
	%5 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%6 = call i64 @strlen(i8* %5)
	%7 = add i64 %6, 1
	%8 = call i8* @malloc(i64 %7)
	%9 = call i8* @memcpy(i8* %8, i8* %5, i64 %7)
	ret i8* %8
}

define i64 @IO_in_int(i8* %self) {
0:
	%1 = alloca i64
	%2 = call i32 (i8*, ...) @scanf([5 x i8]* @str.3, i64* %1)
	%3 = call i32 (i8*, ...) @scanf(i8* getelementptr ([4 x i8], [4 x i8]* @str.5, i32 0, i32 0))
	%4 = load i64, i64* %1
	ret i64 %4
}

define i64 @String_length(i8* %self) {
0:
	%1 = call i64 @strlen(i8* %self)
	ret i64 %1
}

define i8* @Printer_print(i8* %self, i8* %x) {
0:
	%1 = alloca i8*
	store i8* %x, i8** %1
	%2 = load i8*, i8** %1
	%3 = call i8* @IO_out_string(i8* %self, i8* %2)
	ret i8* %3
}

define i8* @Printer_println(i8* %self, i8* %x) {
0:
	%1 = alloca i8*
	store i8* %x, i8** %1
	%2 = load i8*, i8** %1
	%3 = call i8* @Printer_print(i8* %self, i8* %2)
	%4 = call i8* @Printer_print(i8* %self, i8* getelementptr ([2 x i8], [2 x i8]* @str.7, i32 0, i32 0))
	ret i8* %4
}

define i8* @Printer_type_name(i8* %self) {
0:
	ret i8* getelementptr ([8 x i8], [8 x i8]* @str.6, i32 0, i32 0)
}

define i8* @Main_main(i8* %self) {
0:
	%1 = alloca i8*
	store i8* getelementptr ([15 x i8], [15 x i8]* @str.9, i32 0, i32 0), i8** %1
	%2 = call i8* @Printer_println(i8* %self, i8* getelementptr ([15 x i8], [15 x i8]* @str.10, i32 0, i32 0))
	%3 = call i8* @Printer_print(i8* %self, i8* getelementptr ([2 x i8], [2 x i8]* @str.11, i32 0, i32 0))
	%4 = load i8*, i8** %1
	%5 = call i8* @Printer_print(i8* %self, i8* %4)
	%6 = call i8* @Printer_print(i8* %self, i8* getelementptr ([3 x i8], [3 x i8]* @str.12, i32 0, i32 0))
	%7 = call i8* @Printer_print(i8* %self, i8* getelementptr ([5 x i8], [5 x i8]* @str.13, i32 0, i32 0))
	%8 = load i8*, i8** %1
	%9 = call i64 @String_length(i8* %8)
	%10 = call i8* @IO_out_int(i8* %self, i64 %9)
	ret i8* %10
}

define i8* @Main_type_name(i8* %self) {
0:
	ret i8* getelementptr ([5 x i8], [5 x i8]* @str.8, i32 0, i32 0)
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
