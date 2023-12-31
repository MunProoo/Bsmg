/************************************************
 * bm_Main.js
 * Created at 2022. 5. 10. 오전 9:48:13.
 *
 * @author SW2Team
 ************************************************/

exports.setMemberInfo = function(dm_memberInfo){
	var dmMemberInfo = app.lookup("dm_memberInfo");
	dmMemberInfo.build(dm_memberInfo);
}

/*
 * 루트 컨테이너에서 load 이벤트 발생 시 호출.
 * 앱이 최초 구성된후 최초 랜더링 직후에 발생하는 이벤트 입니다.
 */
function onBodyLoad(/* cpr.events.CEvent */ e){
	app.lookup("sms_chkLogin").send();
	var mem_rank = app.lookup("dm_memberInfo").getString("mem_rank");	
	if(mem_rank == '관리자'){
		app.lookup("user_regist").visible = true;
		app.lookup("userManagement").visible = true;
	}
	app.getContainer().redraw();
}


/*
 * "사용자 등록" 버튼(user_regist)에서 click 이벤트 발생 시 호출.
 * 사용자가 컨트롤을 클릭할 때 발생하는 이벤트.
 */
function onUser_registClick(/* cpr.events.CMouseEvent */ e){
	/** 
	 * @type cpr.controls.Button
	 */
	var user_regist = e.control;
	app.getRootAppInstance().openDialog("app/Bsmg/bm_regist", {
		width : 800, height : 600
	}, function(dialog){
		dialog.ready(function(dialogApp){
			dialog.modal = true;
			dialog.headerVisible = true;
			dialog.headerClose = true;
			dialog.headerMovable = true;
			dialog.resizable = true;
			dialog.headerTitle = "사용자 등록";
			dialog.addEventListener("keyup", function(e){
				if (e.keyCode == 27){
					dialog.close();
				}
			});
		});
	})
	
}


/*
 * "로그아웃" 버튼(logout)에서 click 이벤트 발생 시 호출.
 * 사용자가 컨트롤을 클릭할 때 발생하는 이벤트.
 */
function onLogoutClick(/* cpr.events.CMouseEvent */ e){
	/** 
	 * @type cpr.controls.Button
	 */
	
	var logout = e.control;
//	console.log(app.lookup("Result").getString("ResultCode"));
	
	if(confirm("로그아웃 하시겠습니까?")){
		app.lookup("sms_logout").send();
		
	}
	
}

/*
 * 서브미션에서 submit-done 이벤트 발생 시 호출.
 * 응답처리가 모두 종료되면 발생합니다.
 */
function onSms_logoutSubmitDone(/* cpr.events.CSubmissionEvent */ e){
	/** 
	 * @type cpr.protocols.Submission
	 */
	var sms_logout = e.control;
	var result = app.lookup("Result").getString("ResultCode");
	
	if(result == 0){
		alert("정상적으로 로그아웃 되었습니다.");
		cpr.core.App.load("app/Bsmg/bm_login", function(newapp){
			app.close();
			var newInst = newapp.createNewInstance();
			newInst.run();
		});
		return; 
	}
}


/*
 * "사용자 관리" 버튼(userManagement)에서 click 이벤트 발생 시 호출.
 * 사용자가 컨트롤을 클릭할 때 발생하는 이벤트.
 */
function onUserManagementClick(/* cpr.events.CMouseEvent */ e){
	/** 
	 * @type cpr.controls.Button
	 */
	var userManagement = e.control;
	app.getRootAppInstance().openDialog("app/Bsmg/bm_userManagement", {
		width : 800, height : 600
	}, function(dialog){
		dialog.ready(function(dialogApp){
			dialog.modal = true;
			dialog.headerVisible = true;
			dialog.headerClose = true;
			dialog.headerMovable = true;
			dialog.resizable = true;
			dialog.headerTitle = "사용자 관리";
			dialog.addEventListener("keyup", function(e){
				if (e.keyCode == 27){
					dialog.close();
				}
			});
		});
	})
}




/*
 * 서브미션에서 submit-done 이벤트 발생 시 호출.
 * 응답처리가 모두 종료되면 발생합니다.
 */
function onSms_chkLoginSubmitDone(/* cpr.events.CSubmissionEvent */ e){
	/** 
	 * @type cpr.protocols.Submission
	 */
	var sms_chkLogin = e.control;
	var result = app.lookup("Result").getString("ResultCode");
	if(result != 0){
		alert("세션이 끊어졌습니다.");
		app.close();
	}
}
