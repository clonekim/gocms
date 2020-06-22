var PATTERNS = {}


function ynFormat(value, row, index) {
    return (value === 'y' || value === true || value === 1 || value === '1') ? '예' : (value === 'n' || value === false || value === 0 || value === '0') ? '아니오' : 'N/A';
}


function itemsFormat(value, row, index) {
    if (value)
        return value.split(',').join('<br/>');
    else
        return '';
}

function brFormat(value, row, index) {
    if (value)
        return value.split(',').join('<br/>');
    else
        return '';
}


function dateShortFormat(value) {
    if (value)
        return dateFns.format(value, 'YYYY/MM/DD');
    else
        return '';
}

function dateLongFormat(value) {
    if (value)
        return dateFns.format(value, 'YYYY-MM-DD HH:mm');
    else
        return '';
}



function h2html(value) {
    return $('<span>').append($('<h2>').text(value)).html();
}


function addTab(url, title) {
    if ($('#easyui-main-tabs').tabs('exists', title)) {
        $('#easyui-main-tabs').tabs('select', title)
    } else {
        $('#easyui-main-tabs').tabs('add', {
            title: title,
            href: url,
            closable: true,
            cache: false
        });
    }
}

function closeTab(title) {
    $('#easyui-main-tabs').tabs('close', title);
}

function gridOnLoadSuccess(res) {
    $(this).datagrid('getPager').pagination({
        total: res.total,
        pageSize: res.limit,
        pageNumber: res.page
    });
}

function formOnLoadError() {
    alert('데이터를 불러오는데 실패하였습니다');
}

function closeDialog(id) {
    $(id).find('form').form('clear');
    $(id).dialog('close');
}

function toErrorMap(res) {
    var json = $.parseJSON(res);
    alert($.map(json.errors || {}, function (err) {
        return err[0]
    }).join('\r\n'));
}

//jQuery
jQuery.fn.serializeJson = function () {
    var obj = {};
    try {
        var arr = this.serializeArray();
        if (arr) {
            jQuery.each(arr, function () {
                obj[this.name] = this.value;
            });
        }
    } catch (e) {
        alert(e.message);
    }
    return obj;
}

$.fn.datebox.defaults.formatter = function (date) {
    return dateFns.format(date, 'YYYY/MM/DD');
}

$.fn.datebox.defaults.parser = function (s) {
    var t = Date.parse(s);
    if (!isNaN(t))
        return new Date(t);
    else
        return new Date();
}


var easyUIToggle = {
    show: function (jq) {
        $(jq).parent().show();
    },
    hide: function (jq) {
        $(jq).parent().hide();
    }
};
// Extends easyUI method
$.extend($.fn.combobox.methods, easyUIToggle);
$.extend($.fn.textbox.methods, easyUIToggle);
$.extend($.fn.combogrid.methods, easyUIToggle);
$.fn.cleanify = function () {
    this.children().each(function (i, e) {
        if ($(e).attr('textboxname')) {
            $(e).textbox('setValue', '');
        } else if ($(e).attr('comboname')) {
            $(e).combobox('setValue', '');
        }
    });
}


function rgb2Num(value) {
    if (!value)
        return '';
    var hex = value.replace('#', '');
    return parseInt(hex.substring(0, 2), 16) + ',' + parseInt(hex.substring(2, 4), 16) + ',' + parseInt(hex.substring(4, 6), 16);

}

function fileSizeFormat(bytes) {
    if (!bytes)
        return 'N/A';
    var i = -1;
    var byteUnits = [' kB', ' MB', ' GB', ' TB', 'PB', 'EB', 'ZB', 'YB'];
    do {
        bytes = bytes / 1024;
        i++;
    } while (bytes > 1024);
    return Math.max(bytes, 0.1).toFixed(2) + byteUnits[i];
}

function templateFormat() {
    var args = Array.prototype.slice.call(arguments),
        text = args.shift();
    return text.replace(/\{(\d+)\}/g, function (match, key) {
        return typeof args[key] !== 'undefined' ? args[key] : match;
    })
}

//Axios Interceptors
function axiosErrorResponse(error) {
    if (_.isString(error.response.data.errors)) {
        return error.response.data.errors;
    } else if (_.isArray(error.response.data.errors)) {
        return error.response.data.errors.join('\n');
    } else if (_.isObject(_.get(error.response, 'data.errors', error.response.data))) {
        var errors = _.get(error.response, 'data.errors', error.response.data);
        var errs = _.map(errors, function (i, j) {
            return _.isArray(i) ? i[0] : i;
        });
        return errs.join('\n');
    } else {
        return error.response.statusText;
    }
}

axios.interceptors.response.use(function (response) {
    return response;
}, function (error) {
    if (!_.get(error, 'response')) {
        return Promise.reject(error);
    }
    if (error.response.status === 400 || error.response.status === 500) {
        alert(axiosErrorResponse(error));
    }
    return Promise.reject(error);
})