%{
package dsl
%}

%union{
	expr  Expression
	value string
	intValue int
}

%type <expr> root
%type <expr> expr

%token <value> NF_FIELD STRING AND OR EQ NEQ GT LT GE LE REG NREG OPEN_PARENTHESIS CLOSE_PARENTHESIS WITH WITHOUT
%token <intValue> NUMBER
%left AND
%left OR
%%

root:
        expr {
		$$ = $1
		yylex.(*Lexer).result = $$
	}

expr:
	OPEN_PARENTHESIS expr CLOSE_PARENTHESIS { $$ = ParenthesisExpr{inner: $2} }
	| expr AND expr { $$ = LogicalExpr{left: $1, operator: operatorAnd, right: $3} }
	| expr OR expr { $$ = LogicalExpr{left: $1, operator: operatorOr, right: $3} }
	| WITH OPEN_PARENTHESIS NF_FIELD CLOSE_PARENTHESIS { $$ = WithExpr{key: $3} }
	| WITHOUT OPEN_PARENTHESIS NF_FIELD CLOSE_PARENTHESIS { $$ = WithoutExpr{key: $3} }
	| NF_FIELD EQ STRING { $$ = EqExpr{key: $1, value: $3} }
	| NF_FIELD NEQ STRING { $$ = NEqExpr{key: $1, value: $3} }
	| NF_FIELD EQ NUMBER { $$ = EqNumExpr{key: $1, value: $3} }
	| NF_FIELD NEQ NUMBER { $$ = NEqNumExpr{key: $1, value: $3} }
	| NF_FIELD LT NUMBER { $$ = LessThanExpr{key: $1, value: $3} }
	| NF_FIELD GT NUMBER { $$ = GreaterThanExpr{key: $1, value: $3} }
	| NF_FIELD LE NUMBER { $$ = LessOrEqualThanExpr{key: $1, value: $3} }
	| NF_FIELD GE NUMBER { $$ = GreaterOrEqualThanExpr{key: $1, value: $3} }
	| NF_FIELD REG STRING { $$ = RegExpr{key: $1, value: $3} }
	| NF_FIELD NREG STRING { $$ = NRegExpr{key: $1, value: $3} }
%%
