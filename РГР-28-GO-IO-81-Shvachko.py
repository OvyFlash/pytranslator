def toLowerCase(sym): #function that changes uppercase to lowercase symbol
    return sym+32 if 65 <= sym and sym <= 90 else sym #if symbol is in range [65; 90] we should add 32 to make lowercase symbol

def main(): #main function that will write ASCII of uppercase symbol and return ASCII of uppercase
    symbol = 66 #code of uppercase symbol
    symbol = toLowerCase(symbol) #call function that gets uppercase symbol
    return symbol #return code of uppercase symbol
