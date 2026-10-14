package deployment

import "testing"

func TestConfigValidate(t *testing.T){
 for _,tc:=range []struct{mode string;valid bool}{{"",true},{Cloud,true},{SelfHosted,true},{"invalid",false}}{
  err:=(Config{Mode:tc.mode}).Validate()
  if (err==nil)!=tc.valid {t.Errorf("Validate(%q) = %v",tc.mode,err)}
 }
}
