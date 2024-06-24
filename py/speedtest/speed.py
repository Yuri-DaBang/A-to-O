############################
###                      ###
### THE GRAND SPEED TEST ### 
###                      ###
############################

### TEST 1 ###

import datetime, os

def test1(Mrange):
    st = datetime.datetime.now()
    print(st)
    tlist = []    

    for i in range(1,Mrange):
        tlist.append(i)
        #print(i)
        #os.system("cls")

    et = datetime.datetime.now()
    print(et)
    print(f"[*1] Process Done!: {Mrange} DONE: Time taken: {et-st}")
    return et-st
    
def test2(Mrange):
    st = datetime.datetime.now()
    print(st)
    tlist = 0   

    for i in range(1,Mrange):
        tlist += 1
        #print(i)
        #os.system("cls")

    et = datetime.datetime.now()
    print(et)
    print(f"[*2] Process Done!: {Mrange} DONE: Time taken: {et-st}")
    return et-st

# phase2

tol_tst1 = []
tol_tst2 = []

for _ in range(1):
    tol_tst1.append(test1(1000000))
    #tol_tst2.append(1)#test2(10000000))

tol_tst1_val = 0
#tol_tst2_val = 0

for _ in range(len(tol_tst1)):

    sel_valtest1 = str(tol_tst1[_])
    #sel_valtest2 = str(tol_tst2[_])
    
    # float extraction
    valtst1str = sel_valtest1.split(":")
    #valtst2str = sel_valtest2.split(":")

    tol_tst1_val += float(valtst1str[2])
    #tol_tst2_val += float(valtst2str[2])
    
print(f"[**] Test1: Averedge: {tol_tst1_val/len(tol_tst1)}")#\n     Test2: Averedge: {tol_tst2_val/len(tol_tst2)}")